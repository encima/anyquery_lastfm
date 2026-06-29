package main

import (
	"encoding/json"
	"testing"
)

func TestRecentTracksResponseParsing(t *testing.T) {
	payload := `{
		"recenttracks": {
			"track": [
				{
					"artist": {"#text": "Radiohead", "mbid": "a74b1b7f-71a5-4011-9441-d0b5e4122711"},
					"name": "Creep",
					"album": {"#text": "Pablo Honey"},
					"url": "https://www.last.fm/music/Radiohead/_/Creep",
					"date": {"uts": "1700000000", "#text": "14 Nov 2023, 13:20"},
					"@attr": {"nowplaying": "true"},
					"loved": "0",
					"mbid": "track-mbid",
					"image": [{"#text": "", "size": "small"}, {"#text": "https://large.png", "size": "large"}]
				}
			],
			"@attr": {"user": "testuser", "page": "1", "perPage": "200", "totalPages": "1", "total": "1"}
		}
	}`

	var resp recentTracksResponse
	if err := json.Unmarshal([]byte(payload), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(resp.RecentTracks.Track) != 1 {
		t.Fatalf("expected 1 track, got %d", len(resp.RecentTracks.Track))
	}
	tr := resp.RecentTracks.Track[0]
	if tr.Artist.Text != "Radiohead" {
		t.Errorf("expected artist Radiohead, got %s", tr.Artist.Text)
	}
	if tr.Name != "Creep" {
		t.Errorf("expected name Creep, got %s", tr.Name)
	}
	if tr.Attr == nil || tr.Attr.NowPlaying != "true" {
		t.Error("expected now playing true")
	}
	if tr.Date == nil || tr.Date.UTS != "1700000000" {
		t.Errorf("expected date uts 1700000000, got %v", tr.Date)
	}
	if imageURL(tr.Images) != "https://large.png" {
		t.Errorf("expected large image url, got %s", imageURL(tr.Images))
	}
}

func TestTopArtistsResponseParsing(t *testing.T) {
	payload := `{
		"topartists": {
			"artist": [
				{
					"name": "Radiohead",
					"playcount": "1234",
					"listeners": "4567890",
					"url": "https://www.last.fm/music/Radiohead",
					"mbid": "mbid",
					"image": [{"#text": "https://large.png", "size": "large"}],
					"@attr": {"rank": "1"}
				}
			],
			"@attr": {"user": "testuser", "page": "1", "perPage": "200", "totalPages": "1", "total": "1"}
		}
	}`

	var resp topArtistsResponse
	if err := json.Unmarshal([]byte(payload), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(resp.TopArtists.Artist) != 1 {
		t.Fatalf("expected 1 artist, got %d", len(resp.TopArtists.Artist))
	}
	ar := resp.TopArtists.Artist[0]
	if ar.Attr.Rank != "1" {
		t.Errorf("expected rank 1, got %s", ar.Attr.Rank)
	}
	if parseInt(ar.PlayCount) != 1234 {
		t.Errorf("expected playcount 1234, got %d", parseInt(ar.PlayCount))
	}
}

func TestMaxPagesForLimit(t *testing.T) {
	cases := []struct {
		limit   int
		perPage int
		want    int
	}{
		{0, 200, 0},
		{-1, 200, 0},
		{10, 200, 1},
		{200, 200, 1},
		{201, 200, 2},
		{400, 200, 2},
	}
	for _, tc := range cases {
		got := maxPagesForLimit(tc.limit, tc.perPage)
		if got != tc.want {
			t.Errorf("maxPagesForLimit(%d, %d) = %d, want %d", tc.limit, tc.perPage, got, tc.want)
		}
	}
}
