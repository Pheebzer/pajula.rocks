package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func getAccessToken(c *http.Client, url string, key string) string {
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

	var j map[string]interface{}
	json.NewDecoder(res.Body).Decode(&j)
	if err != nil {
		panic(err)
	}

	return j["access_token"].(string)
}

type Track struct {
	User     string
	Date     string
	Name     string
	Artist   string
	Album    string
	Duration int
}
