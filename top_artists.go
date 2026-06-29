package main

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/julien040/anyquery/rpc"
)

const topArtistsPerPage = 200

// topArtistsTable reads a user's top artists.
type topArtistsTable struct {
	client *lastFMClient
}

func topArtistsCreator(args rpc.TableCreatorArgs) (rpc.Table, *rpc.DatabaseSchema, error) {
	apiKey := args.UserConfig.GetString("api_key")
	if apiKey == "" {
		return nil, nil, fmt.Errorf("api_key must be set in the plugin configuration")
	}

	return &topArtistsTable{
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
				Description: "Rank in the user's top artists",
			},
			{
				Name:        "name",
				Type:        rpc.ColumnTypeString,
				Description: "Artist name",
			},
			{
				Name:        "playcount",
				Type:        rpc.ColumnTypeInt,
				Description: "Number of plays by the user",
			},
			{
				Name:        "listeners",
				Type:        rpc.ColumnTypeInt,
				Description: "Global number of listeners",
			},
			{
				Name:        "url",
				Type:        rpc.ColumnTypeString,
				Description: "Last.fm URL for the artist",
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

func (t *topArtistsTable) CreateReader() rpc.ReaderInterface {
	return &topArtistsCursor{
		table: t,
		page:  1,
	}
}

func (t *topArtistsTable) Close() error {
	return nil
}

type topArtistsCursor struct {
	table *topArtistsTable
	page  int
}

func (c *topArtistsCursor) Query(constraints rpc.QueryConstraint) ([][]interface{}, bool, error) {
	user := constraints.GetColumnConstraint(0).GetStringValue()
	if user == "" {
		return nil, true, fmt.Errorf("user parameter is required")
	}

	params := url.Values{}
	params.Set("user", user)
	params.Set("limit", strconv.Itoa(topArtistsPerPage))
	params.Set("page", strconv.Itoa(c.page))

	var resp topArtistsResponse
	if err := c.table.client.fetch("user.gettopartists", params, &resp); err != nil {
		return nil, true, err
	}

	rows := make([][]interface{}, 0, len(resp.TopArtists.Artist))
	for _, ar := range resp.TopArtists.Artist {
		rows = append(rows, []interface{}{
			parseInt(ar.Attr.Rank),
			ar.Name,
			parseInt(ar.PlayCount),
			parseInt(ar.Listeners),
			ar.URL,
			ar.MBID,
			imageURL(ar.Images),
		})
	}

	totalPages := 0
	if resp.TopArtists.Attr != nil {
		totalPages = int(parseInt(resp.TopArtists.Attr.TotalPages))
	}

	maxPages := maxPagesForLimit(constraints.Limit, topArtistsPerPage)
	c.page++
	noMore := c.page > totalPages || len(resp.TopArtists.Artist) == 0
	if maxPages > 0 && c.page > maxPages {
		noMore = true
	}
	return rows, noMore, nil
}
