package main

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/julien040/anyquery/rpc"
)

const topTracksPerPage = 200

// topTracksTable reads a user's top tracks.
type topTracksTable struct {
	client *lastFMClient
}

func topTracksCreator(args rpc.TableCreatorArgs) (rpc.Table, *rpc.DatabaseSchema, error) {
	apiKey := args.UserConfig.GetString("api_key")
	if apiKey == "" {
		return nil, nil, fmt.Errorf("api_key must be set in the plugin configuration")
	}

	return &topTracksTable{
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
				Description: "Rank in the user's top tracks",
			},
			{
				Name:        "name",
				Type:        rpc.ColumnTypeString,
				Description: "Track title",
			},
			{
				Name:        "artist",
				Type:        rpc.ColumnTypeString,
				Description: "Track artist",
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
				Description: "Last.fm URL for the track",
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

func (t *topTracksTable) CreateReader() rpc.ReaderInterface {
	return &topTracksCursor{
		table: t,
		page:  1,
	}
}

func (t *topTracksTable) Close() error {
	return nil
}

type topTracksCursor struct {
	table *topTracksTable
	page  int
}

func (c *topTracksCursor) Query(constraints rpc.QueryConstraint) ([][]interface{}, bool, error) {
	user := constraints.GetColumnConstraint(0).GetStringValue()
	if user == "" {
		return nil, true, fmt.Errorf("user parameter is required")
	}

	params := url.Values{}
	params.Set("user", user)
	params.Set("limit", strconv.Itoa(topTracksPerPage))
	params.Set("page", strconv.Itoa(c.page))

	var resp topTracksResponse
	if err := c.table.client.fetch("user.gettoptracks", params, &resp); err != nil {
		return nil, true, err
	}

	rows := make([][]interface{}, 0, len(resp.TopTracks.Track))
	for _, tr := range resp.TopTracks.Track {
		artist := tr.Artist.Text
		if artist == "" {
			artist = tr.Artist.Name
		}
		rows = append(rows, []interface{}{
			parseInt(tr.Attr.Rank),
			tr.Name,
			artist,
			parseInt(tr.PlayCount),
			parseInt(tr.Listeners),
			tr.URL,
			tr.MBID,
			imageURL(tr.Images),
		})
	}

	totalPages := 0
	if resp.TopTracks.Attr != nil {
		totalPages = int(parseInt(resp.TopTracks.Attr.TotalPages))
	}

	maxPages := maxPagesForLimit(constraints.Limit, topTracksPerPage)
	c.page++
	noMore := c.page > totalPages || len(resp.TopTracks.Track) == 0
	if maxPages > 0 && c.page > maxPages {
		noMore = true
	}
	return rows, noMore, nil
}
