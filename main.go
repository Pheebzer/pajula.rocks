package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func rollBackTx(tx *sql.Tx) {
	fmt.Println("Attempting to roll back tx")
	err := tx.Rollback()
	if err != nil {
		fmt.Println(err)
	}
	panic("importer failed")
}

func main() {
	cfp := os.Getenv("CONFIG_FILE")
	if cfp == "" {
		cfp = "config.yaml"
	}
	cfg := parseConfig(cfp)

	// initialise database and start new transaction
	db, err := InitDB(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// initialise http client and related data
	c := &http.Client{}
	token := fetchAccessToken(c, cfg.Spotify.TokenEndpoint, cfg.Spotify.ApiKey)
	filterStr := "fields=next,items(added_by.id,added_at,track(name,id,duration_ms,album(name),artists(name))"
	pageBaseUrl := fmt.Sprintf("%s/%s/tracks", cfg.Spotify.PlaylistEndpoint, cfg.Spotify.PlaylistId)
	snapshotUrl := fmt.Sprintf("%s/%s?fields=snapshot_id", cfg.Spotify.PlaylistEndpoint, cfg.Spotify.PlaylistId)
	nextPage := fmt.Sprintf("%s?%s&limit=100", pageBaseUrl, filterStr)

	// start new transaction, prepare statements
	tx, err := db.Begin()
	if err != nil {
		panic(err)
	}
	getSnapshotId, err := tx.Prepare(getSnapshotQuery)
	if err != nil {
		panic(err)
	}
	updateSnapshotId, err := tx.Prepare(updateSnapshotQuery)
	if err != nil {
		panic(err)
	}
	insertTrack, err := tx.Prepare(insertTrackQuery)
	if err != nil {
		panic(err)
	}
	deleteTracks, err := tx.Prepare(deleteTracksQuery)
	if err != nil {
		panic(err)
	}

	// check if the playlist has changed, return early if there is nothing to import
	newSnapshotId := fetchSnapshotId(c, snapshotUrl, token)
	var snapshotId string
	if err := getSnapshotId.QueryRow().Scan(&snapshotId); err != nil {
		panic(err)
	}
	if newSnapshotId == snapshotId {
		fmt.Println("Snapshot IDs match, nothing to import")
		return
	}

	// update snapshot_id
	_, err = updateSnapshotId.Exec(newSnapshotId)
	if err != nil {
		fmt.Println("unable to update snapshot id")
		panic(err)
	}

	// delete old songs. TRUNCATE not used as it doesn't support transactions
	//
	// NOTE:
	// Spotify playlist max song is 10k -> worst case for deleting + inserting is few seconds
	// refactor this if multiple playlists need to be tracked
	_, err = deleteTracks.Exec()
	if err != nil {
		fmt.Println("unable to delete old tracks")
		panic(err)
	}

	// insert new tracks into the database
	for nextPage != "" {
		pg := fetchPageData(c, nextPage, token)
		for _, e := range pg.Items {
			date, _ := time.Parse(time.RFC3339, e.AddedAt)
			_, err = insertTrack.Exec(
				e.Track.ID,
				e.Track.Name,
				e.Track.Album.Name,
				e.Track.Artists[0].Name,
				e.AddedBy.ID,
				date,
				e.Track.DurationMs,
				snapshotId,
			)
			if err != nil {
				fmt.Printf("Failed to insert new track: %s \n", err)
				rollBackTx(tx)
			}
		}
		nextPage = pg.Next
	}

	err = tx.Commit()
	if err != nil {
		rollBackTx(tx)
	}
	fmt.Println("Done!")
}
