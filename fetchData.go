package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

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
