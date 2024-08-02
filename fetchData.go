package main

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"net/http"
	"strings"
)

type Track struct {
	Name       string
	Album      string
	AddedBy    string
	AddedAt    string
	Hash       uint32
	DurationMs int
}

type PageData struct {
	Items []struct {
		Track struct {
			Artists []struct {
				Name string `json:"name"`
			} `json:"artists"`
			Album struct {
				Name string `json:"name"`
			} `json:"album"`
			DurationMs int    `json:"duration_ms"`
			Name       string `json:"name"`
		} `json:"track"`
		AddedBy struct {
			ID string `json:"id"`
		} `json:"added_by"`
		AddedAt string `json:"added_at"`
	} `json:"items"`
	Next string `json:"next"`
}

func calculateSongHash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func getAccessToken(c *http.Client, url, key string) string {

	req, err := http.NewRequest(
		"POST",
		url,
		strings.NewReader("grant_type=client_credentials"),
	)
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", key))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		panic(err)
	}

	res, err := c.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	j := struct {
		AccessToken string `json:"access_token"`
	}{}
	json.NewDecoder(res.Body).Decode(&j)
	if err != nil {
		panic(err)
	}
	return j.AccessToken
}

func getSnapshotId(c *http.Client, url, token string) string {
	req, err := http.NewRequest(
		"GET",
		url,
		nil,
	)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	if err != nil {
		panic(err)
	}
	res, err := c.Do(req)
	if err != nil {
		panic(err)
	}
	j := struct {
		SnapshotId string `json:"snapshot_id"`
	}{}
	json.NewDecoder(res.Body).Decode(&j)
	if err != nil {
		panic(err)
	}
	return j.SnapshotId
}

func getPageData(c *http.Client, url, token string) *PageData {

	req, err := http.NewRequest(
		"GET",
		url,
		nil,
	)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	if err != nil {
		panic(err)
	}
	res, err := c.Do(req)
	if err != nil {
		panic(err)
	}
	d := PageData{}
	json.NewDecoder(res.Body).Decode(&d)
	if err != nil {
		panic(err)
	}
	return &d
}

func getPlaylistData(cfg *Config) ([]Track, bool) {
	c := &http.Client{}

	token := getAccessToken(c, cfg.Spotify.TokenEndpoint, cfg.Spotify.ApiKey)
	filterStr := "fields=next,items(added_by.id,added_at,track(name,duration_ms,album(name),artists(name))"
	pageBaseUrl := fmt.Sprintf("%s/%s/tracks", cfg.Spotify.PlaylistEndpoint, cfg.Spotify.PlaylistId)

	// check if playlist has changed
	// @TODO: actually pull the previous snapshot_id from the database
	// sidUrl := fmt.Sprintf("%s/%s?fields=snapshot_id", pageBaseUrl, cfg.Spotify.PlaylistId)
	// nsid := getSnapshotId(c, sidUrl, token)
	// sid := "previous sid from database"
	// if sid == nsid {
	// 	return []Track{}, false
	// }

	// Fetch pages until response 'next url' parameter indicates we are done
	next := fmt.Sprintf("%s?%s&limit=100", pageBaseUrl, filterStr)
	masterData := []Track{}
	for next != "" {
		pd := getPageData(c, next, token)
		for _, d := range pd.Items {
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
		next = pd.Next
	}
	return masterData, true
}
