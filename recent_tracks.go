package main

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/julien040/anyquery/rpc"
)

const recentTracksPerPage = 200

// recentTracksTable reads a user's recently scrobbled tracks.
type recentTracksTable struct {
	client *lastFMClient
	user   string
}

func recentTracksCreator(args rpc.TableCreatorArgs) (rpc.Table, *rpc.DatabaseSchema, error) {
	apiKey := args.UserConfig.GetString("api_key")
	if apiKey == "" {
		return nil, nil, fmt.Errorf("api_key must be set in the plugin configuration")
	}

	return &recentTracksTable{
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
				Name:        "artist",
				Type:        rpc.ColumnTypeString,
				Description: "Name of the artist",
			},
			{
				Name:        "name",
				Type:        rpc.ColumnTypeString,
				Description: "Track title",
			},
			{
				Name:        "album",
				Type:        rpc.ColumnTypeString,
				Description: "Album title",
			},
			{
				Name:        "url",
				Type:        rpc.ColumnTypeString,
				Description: "Last.fm URL for the track",
			},
			{
				Name:        "played_at",
				Type:        rpc.ColumnTypeString,
				Description: "When the track was scrobbled (RFC3339, e.g. 2023-11-14T22:13:20Z)",
			},
			{
				Name:        "now_playing",
				Type:        rpc.ColumnTypeBool,
				Description: "Whether the track is currently playing",
			},
			{
				Name:        "loved",
				Type:        rpc.ColumnTypeBool,
				Description: "Whether the track is loved by the user",
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

func (t *recentTracksTable) CreateReader() rpc.ReaderInterface {
	return &recentTracksCursor{
		table: t,
		page:  1,
	}
}

func (t *recentTracksTable) Close() error {
	return nil
}

type recentTracksCursor struct {
	table *recentTracksTable
	page  int
}

func (c *recentTracksCursor) Query(constraints rpc.QueryConstraint) ([][]interface{}, bool, error) {
	user := constraints.GetColumnConstraint(0).GetStringValue()
	if user == "" {
		return nil, true, fmt.Errorf("user parameter is required")
	}

	params := url.Values{}
	params.Set("user", user)
	params.Set("limit", strconv.Itoa(recentTracksPerPage))
	params.Set("page", strconv.Itoa(c.page))

	var resp recentTracksResponse
	if err := c.table.client.fetch("user.getrecenttracks", params, &resp); err != nil {
		return nil, true, err
	}

	rows := make([][]interface{}, 0, len(resp.RecentTracks.Track))
	for _, tr := range resp.RecentTracks.Track {
		artist := tr.Artist.Text
		if artist == "" {
			artist = tr.Artist.Name
		}

		var playedAt interface{}
		if tr.Date != nil && tr.Date.UTS != "" {
			if uts := parseInt(tr.Date.UTS); uts > 0 {
				playedAt = time.Unix(uts, 0).UTC().Format(time.RFC3339)
			}
		}

		nowPlaying := tr.Attr != nil && tr.Attr.NowPlaying == "true"
		loved := tr.Loved == "1"

		rows = append(rows, []interface{}{
			artist,
			tr.Name,
			tr.Album.Text,
			tr.URL,
			playedAt,
			nowPlaying,
			loved,
			tr.MBID,
			imageURL(tr.Images),
		})
	}

	totalPages := 0
	if resp.RecentTracks.Attr != nil {
		totalPages = int(parseInt(resp.RecentTracks.Attr.TotalPages))
	}

	maxPages := maxPagesForLimit(constraints.Limit, recentTracksPerPage)
	c.page++
	noMore := c.page > totalPages || len(resp.RecentTracks.Track) == 0
	if maxPages > 0 && c.page > maxPages {
		noMore = true
	}
	return rows, noMore, nil
}
