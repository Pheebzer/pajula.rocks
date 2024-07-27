package main

import (
	"net/http"
)

func main() {
	cfg := parseConfig("config.yaml")
	c := &http.Client{}
	token := getAccessToken(c, cfg.Spotify.TokenEndpoint, cfg.Spotify.ApiKey)

	next := "https://api.spotify.com/v1/playlists/6KpVVH6Njm4jnZmKdQBQuI/tracks?limit=100&fields=next,items(added_by.id,added_at,track(name,album(name),artists(name,duration_ms))"
	for next != "" {

	}

}
