package main

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/julien040/anyquery/rpc"
)

const topAlbumsPerPage = 200

// topAlbumsTable reads a user's top albums.
type topAlbumsTable struct {
	client *lastFMClient
}

func topAlbumsCreator(args rpc.TableCreatorArgs) (rpc.Table, *rpc.DatabaseSchema, error) {
	apiKey := args.UserConfig.GetString("api_key")
	if apiKey == "" {
		return nil, nil, fmt.Errorf("api_key must be set in the plugin configuration")
	}

	return &topAlbumsTable{
		client: newLastFMClient(apiKey),
	}, &rpc.DatabaseSchema{
		Columns: []rpc.DatabaseSchemaColumn{
			{
				Name:        "user",
				Type:        rpc.ColumnTypeString,
				IsParameter: true,
				IsRequired:  true,
				Description: "The Last.fm username",
			},
			{
				Name:        "rank",
				Type:        rpc.ColumnTypeInt,
				Description: "Rank in the user's top albums",
			},
			{
				Name:        "name",
				Type:        rpc.ColumnTypeString,
				Description: "Album title",
			},
			{
				Name:        "artist",
				Type:        rpc.ColumnTypeString,
				Description: "Album artist",
			},
			{
				Name:        "playcount",
				Type:        rpc.ColumnTypeInt,
				Description: "Number of plays by the user",
			},
			{
				Name:        "url",
				Type:        rpc.ColumnTypeString,
				Description: "Last.fm URL for the album",
			},
			{
				Name:        "mbid",
				Type:        rpc.ColumnTypeString,
				Description: "MusicBrainz ID",
			},
			{
				Name:        "image",
				Type:        rpc.ColumnTypeString,
				Description: "URL of the largest available image",
			},
		},
		PrimaryKey: -1,
	}, nil
}

func (t *topAlbumsTable) CreateReader() rpc.ReaderInterface {
	return &topAlbumsCursor{
		table: t,
		page:  1,
	}
}

func (t *topAlbumsTable) Close() error {
	return nil
}

type topAlbumsCursor struct {
	table *topAlbumsTable
	page  int
}

func (c *topAlbumsCursor) Query(constraints rpc.QueryConstraint) ([][]interface{}, bool, error) {
	user := constraints.GetColumnConstraint(0).GetStringValue()
	if user == "" {
		return nil, true, fmt.Errorf("user parameter is required")
	}

	params := url.Values{}
	params.Set("user", user)
	params.Set("limit", strconv.Itoa(topAlbumsPerPage))
	params.Set("page", strconv.Itoa(c.page))

	var resp topAlbumsResponse
	if err := c.table.client.fetch("user.gettopalbums", params, &resp); err != nil {
		return nil, true, err
	}

	rows := make([][]interface{}, 0, len(resp.TopAlbums.Album))
	for _, al := range resp.TopAlbums.Album {
		artist := al.Artist.Text
		if artist == "" {
			artist = al.Artist.Name
		}
		rows = append(rows, []interface{}{
			parseInt(al.Attr.Rank),
			al.Name,
			artist,
			parseInt(al.PlayCount),
			al.URL,
			al.MBID,
			imageURL(al.Images),
		})
	}

	totalPages := 0
	if resp.TopAlbums.Attr != nil {
		totalPages = int(parseInt(resp.TopAlbums.Attr.TotalPages))
	}

	maxPages := maxPagesForLimit(constraints.Limit, topAlbumsPerPage)
	c.page++
	noMore := c.page > totalPages || len(resp.TopAlbums.Album) == 0
	if maxPages > 0 && c.page > maxPages {
		noMore = true
	}
	return rows, noMore, nil
}
