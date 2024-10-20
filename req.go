package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type TokenData struct {
	AccessToken string `json:"access_token"`
}
type MetaData struct {
	Tracks struct {
		TrackCount int `json:"total"`
	} `json:"tracks"`
	SnapshotId string `json:"snapshot_id"`
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
			ID         string `json:"id"`
		} `json:"track"`
		AddedBy struct {
			ID string `json:"id"`
		} `json:"added_by"`
		AddedAt string `json:"added_at"`
	} `json:"items"`
	Next string `json:"next"`
}
type Track struct {
	Id         string
	Name       string
	Album      string
	Artist     string
	AddedBy    string
	AddedAt    time.Time
	DurationMs int
	SnapshotId string
}

func fetchAccessToken(c *http.Client, url, key string) TokenData {
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

	var j TokenData
	json.NewDecoder(res.Body).Decode(&j)
	if err != nil {
		panic(err)
	}
	return j
}

func fetchMetadata(c *http.Client, url, token string) (snapshotId string, songCount int) {
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
	var j MetaData
	json.NewDecoder(res.Body).Decode(&j)
	if err != nil {
		panic(err)
	}
	return j.SnapshotId, j.Tracks.TrackCount
}

func fetchPageData(c *http.Client, url, token, snid string) []Track {
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
	var t []Track
	json.NewDecoder(res.Body).Decode(&d)
	if err != nil {
		panic(err)
	}
	for _, e := range d.Items {
		at, _ := time.Parse(time.RFC3339, e.AddedAt)
		t = append(t, Track{
			Id:         e.Track.ID,
			Name:       e.Track.Name,
			Album:      e.Track.Album.Name,
			Artist:     e.Track.Artists[0].Name,
			AddedBy:    e.AddedBy.ID,
			AddedAt:    at,
			DurationMs: e.Track.DurationMs,
			SnapshotId: snid,
		})
	}
	return t
}
