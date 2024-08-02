package main

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	cfp := os.Getenv("CONFIG_FILE")
	if cfp == "" {
		cfp = "config.yaml"
	}
	cfg := parseConfig(cfp)

	pd, changed := getPlaylistData(cfg)
	if !changed {
		fmt.Println("No changes to import")
		return
	}

}
