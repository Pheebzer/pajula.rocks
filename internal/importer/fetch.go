package importer

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	logger "pajula.rocks/internal/log"
)

var httpc = &http.Client{
	Timeout: 10 * time.Second,
}

type TokenData struct {
	AccessToken string `json:"access_token"`
}
type MetaData struct {
	Tracks struct {
		TrackCount int `json:"total"`
	} `json:"tracks"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SnapshotId  string `json:"snapshot_id"`
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

func doRequest(req *http.Request) *http.Response {
	res, err := httpc.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		logger.Fatalf("Invalid API response %d for request %s", res.StatusCode, req.URL)
	}
	return res
}

func FetchAccessToken(url, key string) TokenData {
	req, err := http.NewRequest(
		"POST",
		url,
		strings.NewReader("grant_type=client_credentials"),
	)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Basic %s", key))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res := doRequest(req)
	defer res.Body.Close()

	var j TokenData
	json.NewDecoder(res.Body).Decode(&j)
	if err != nil {
		log.Fatal(err)
	}

	return j
}

func FetchMetadata(url, token string) MetaData {
	req, err := http.NewRequest(
		"GET",
		url,
		nil,
	)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	if err != nil {
		log.Fatal(err)
	}

	res := doRequest(req)
	defer res.Body.Close()

	var m MetaData
	json.NewDecoder(res.Body).Decode(&m)
	if err != nil {
		log.Fatal(err)
	}

	return m
}

func FetchPageData(url, token, snid string) []Track {
	req, err := http.NewRequest(
		"GET",
		url,
		nil,
	)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	if err != nil {
		log.Fatal(err)
	}

	res := doRequest(req)
	defer res.Body.Close()

	d := PageData{}
	var t []Track
	json.NewDecoder(res.Body).Decode(&d)
	if err != nil {
		log.Fatal(err)
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
