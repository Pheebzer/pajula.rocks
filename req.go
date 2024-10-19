package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type TokenData struct {
	AccessToken string `json:"access_token"`
}
type MetaData struct {
	SnapshotId string `json:"snapshot_id"`
	TrackCount int    `json:"total"`
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
	return j.SnapshotId, j.TrackCount
}

func fetchPageData(c *http.Client, url, token string) *PageData {

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
