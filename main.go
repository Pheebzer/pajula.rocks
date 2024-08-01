package main

import (
	"fmt"
	"hash/fnv"
	"net/http"
)

type Track struct {
	Name       string
	Album      string
	AddedBy    string
	AddedAt    string
	Hash       uint32
	DurationMs int
}

func calculateSongHash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func main() {
	// parse configurations values
	cfg := parseConfig("config.yaml")

	// initialize http client and fetch API access token
	c := &http.Client{}
	token := getAccessToken(c, cfg.Spotify.TokenEndpoint, cfg.Spotify.ApiKey)

	// Build url
	filterStr := "fields=next,items(added_by.id,added_at,track(name,album(name),artists(name,duration_ms))"
	pageBaseUrl := fmt.Sprintf("%s/%s/tracks", cfg.Spotify.PlaylistEndpoint, cfg.Spotify.PlaylistId)
	next := fmt.Sprintf("%s?%s&limit=100", pageBaseUrl, filterStr)
	// Fetch pages until response 'next url' parameter indicates we are done
	masterData := []Track{}
	for next != "" {
		fmt.Printf("Next page: %s", next)
		pageData := getPageData(c, next, token)
		// unroll page data, append to master list
		for _, d := range pageData.Items {
			hs := d.Track.Album.Name + d.Track.Artists[0].Name
			h := calculateSongHash(hs)
			masterData = append(masterData, Track{
				Name:       d.Track.Artists[0].Name,
				Album:      d.Track.Album.Name,
				AddedBy:    d.AddedBy.ID,
				AddedAt:    d.AddedAt,
				DurationMs: d.Track.DurationMs,
				Hash:       h,
			})
		}
		// Fetch next page if exists
		next = pageData.Next
	}
	fmt.Printf("Data fetching done")
	fmt.Println(masterData)
}
