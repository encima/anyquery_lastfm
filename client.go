package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const lastFMBaseURL = "https://ws.audioscrobbler.com/2.0/"

// lastFMClient is a small HTTP client for the Last.fm API.
type lastFMClient struct {
	apiKey string
	client *http.Client
}

// newLastFMClient creates a new Last.fm API client.
func newLastFMClient(apiKey string) *lastFMClient {
	return &lastFMClient{
		apiKey: apiKey,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// fetch calls the Last.fm API with the given method and parameters and
// unmarshals the JSON response into dst.
func (c *lastFMClient) fetch(method string, params url.Values, dst interface{}) error {
	if c.apiKey == "" {
		return fmt.Errorf("last.fm api_key is not configured")
	}

	query := url.Values{}
	query.Set("method", method)
	query.Set("api_key", c.apiKey)
	query.Set("format", "json")
	for k, v := range params {
		query[k] = v
	}

	reqURL := lastFMBaseURL + "?" + query.Encode()
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call last.fm API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read last.fm response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("last.fm API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Last.fm reports API errors as HTTP 200 with an "error" field.
	var apiErr lastFMError
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.ErrorCode != 0 {
		return fmt.Errorf("last.fm API error %d: %s", apiErr.ErrorCode, apiErr.Message)
	}

	if err := json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("failed to decode last.fm response: %w", err)
	}
	return nil
}

// lastFMError represents an error returned by the Last.fm API.
type lastFMError struct {
	ErrorCode int    `json:"error"`
	Message   string `json:"message"`
}

// imageURL returns the URL of the largest image available, or an empty string.
func imageURL(images []lastFMImage) string {
	var url string
	for _, img := range images {
		if img.Text != "" {
			url = img.Text
		}
	}
	return url
}

// parseInt parses a string integer, returning 0 on failure.
func parseInt(s string) int64 {
	if s == "" {
		return 0
	}
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

// maxPagesForLimit returns the maximum number of pages needed to satisfy a
// LIMIT clause, or 0 when no limit is set.
func maxPagesForLimit(limit int, perPage int) int {
	if limit <= 0 {
		return 0
	}
	pages := limit / perPage
	if limit%perPage != 0 {
		pages++
	}
	return pages
}

// Last.fm shared types.

type lastFMImage struct {
	Text string `json:"#text"`
	Size string `json:"size"`
}

type lastFMArtist struct {
	Text string `json:"#text"`
	Name string `json:"name"`
	MBID string `json:"mbid"`
}

type lastFMAlbum struct {
	Text string `json:"#text"`
	Name string `json:"name"`
}

type lastFMDate struct {
	UTS  string `json:"uts"`
	Text string `json:"#text"`
}

// Recent tracks types.

type recentTracksResponse struct {
	RecentTracks recentTracks `json:"recenttracks"`
}

type recentTracks struct {
	Track []recentTrack `json:"track"`
	Attr  *pagerAttr    `json:"@attr"`
}

type recentTrack struct {
	Artist    lastFMArtist  `json:"artist"`
	Name      string        `json:"name"`
	Album     lastFMAlbum   `json:"album"`
	URL       string        `json:"url"`
	Date      *lastFMDate   `json:"date"`
	Attr      *nowPlayingAttr `json:"@attr"`
	Loved     string        `json:"loved"`
	MBID      string        `json:"mbid"`
	Images    []lastFMImage `json:"image"`
}

type nowPlayingAttr struct {
	NowPlaying string `json:"nowplaying"`
}

// Top artists types.

type topArtistsResponse struct {
	TopArtists topArtists `json:"topartists"`
}

type topArtists struct {
	Artist []topArtist `json:"artist"`
	Attr   *pagerAttr  `json:"@attr"`
}

type topArtist struct {
	Name      string        `json:"name"`
	PlayCount string        `json:"playcount"`
	Listeners string        `json:"listeners"`
	URL       string        `json:"url"`
	MBID      string        `json:"mbid"`
	Images    []lastFMImage `json:"image"`
	Attr      artistRankAttr `json:"@attr"`
}

type artistRankAttr struct {
	Rank string `json:"rank"`
}

// Top tracks types.

type topTracksResponse struct {
	TopTracks topTracks `json:"toptracks"`
}

type topTracks struct {
	Track  []topTrack `json:"track"`
	Attr   *pagerAttr `json:"@attr"`
}

type topTrack struct {
	Name      string        `json:"name"`
	Artist    lastFMArtist  `json:"artist"`
	PlayCount string        `json:"playcount"`
	Listeners string        `json:"listeners"`
	URL       string        `json:"url"`
	MBID      string        `json:"mbid"`
	Images    []lastFMImage `json:"image"`
	Attr      trackRankAttr `json:"@attr"`
}

type trackRankAttr struct {
	Rank string `json:"rank"`
}

// Top albums types.

type topAlbumsResponse struct {
	TopAlbums topAlbums `json:"topalbums"`
}

type topAlbums struct {
	Album []topAlbum `json:"album"`
	Attr  *pagerAttr `json:"@attr"`
}

type topAlbum struct {
	Name      string        `json:"name"`
	Artist    lastFMArtist  `json:"artist"`
	PlayCount string        `json:"playcount"`
	URL       string        `json:"url"`
	MBID      string        `json:"mbid"`
	Images    []lastFMImage `json:"image"`
	Attr      albumRankAttr `json:"@attr"`
}

type albumRankAttr struct {
	Rank string `json:"rank"`
}

type pagerAttr struct {
	User       string `json:"user"`
	Page       string `json:"page"`
	PerPage    string `json:"perPage"`
	TotalPages string `json:"totalPages"`
	Total      string `json:"total"`
}
