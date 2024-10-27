package main //32twOqGf8gIswTgzG3IKxP

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func rollBackTx(tx *sql.Tx) {
	log.Println("Attempting to roll back tx")
	err := tx.Rollback()
	if err != nil {
		log.Fatalf("ERROR: Failed to roll back transaction, check data integrity:\n%s", err)
	}
	os.Exit(1)
}

func main() {
	tstart := time.Now()
	initLogs()
	cfg := getConfigs()
	db, err := InitDB(cfg)
	if err != nil {
		log.Fatalf("ERROR: Unable to connect to database:\n%s", err)
	}
	defer db.Close()

	c := &http.Client{}
	token := fetchAccessToken(c, cfg.TokenEndpoint, cfg.ApiKey).AccessToken
	filterStr := "fields=next,items(added_by.id,added_at,track(name,id,duration_ms,album(name),artists(name))"
	pageBaseUrl := fmt.Sprintf("%s/%s/tracks", cfg.PlaylistEndpoint, cfg.PlaylistId)
	metadataUrl := fmt.Sprintf("%s/%s?fields=snapshot_id,tracks.total", cfg.PlaylistEndpoint, cfg.PlaylistId)

	itx, err := StartTransaction(db)
	if err != nil {
		log.Fatal(err)
	}

	// check if the playlist has changed, return early if there is nothing to import
	newSnapshotId, songCount := fetchMetadata(c, metadataUrl, token)
	var oldSnapshotId string
	if err := itx.statements.GetSnapshotId.QueryRow().Scan(&oldSnapshotId); err != nil {
		log.Fatal(err)
	}
	if newSnapshotId == oldSnapshotId {
		log.Println("INFO: Snapshot IDs match, nothing to import")
		os.Exit(0)
		return
	}

	// API rate limit ~180 requests per second, returns max 100 tracks per request
	// playlist max song count is 10,000
	// -> should be possible to always fetch all songs concurrently before rate limit kicks in
	reqCount := int(math.Ceil((float64(songCount) / 100)))
	tracksCh := make(chan []Track, reqCount)
	var wg sync.WaitGroup

	log.Printf("INFO: Fetching songs for playlist %s", cfg.PlaylistId)

	for i := 0; i < reqCount; i++ {
		wg.Add(1)
		apiUrl := fmt.Sprintf("%s?%s&limit=100&offset=%d", pageBaseUrl, filterStr, i*100)
		go func(apiUrl string) {
			defer wg.Done()
			tracks := fetchPageData(c, apiUrl, token, newSnapshotId)
			tracksCh <- tracks
		}(apiUrl)
	}

	wg.Wait()
	close(tracksCh)

	// update snapshot_id
	_, err = itx.statements.UpdateSnapshotId.Exec(newSnapshotId)
	if err != nil {
		log.Fatal(err)
	}

	// Add new tracks, update snapshot_id where applicable
	for ts := range tracksCh {
		for _, t := range ts {
			_, err := itx.statements.InsertUpdateTrack.Exec(
				t.Id,
				t.Name,
				t.Album,
				t.Artist,
				t.AddedBy,
				t.AddedAt,
				t.DurationMs,
				t.SnapshotId,
			)
			if err != nil {
				rollBackTx(itx.tx)
				log.Fatalf("Error updating tracks:\n%s", err)
			}
		}
	}

	// delete old tracks
	_, err = itx.statements.DeleteTracks.Exec(oldSnapshotId)
	if err != nil {
		rollBackTx(itx.tx)
		fmt.Printf("Failed to delete old tracks: %s \n", err)
	}

	err = itx.tx.Commit()
	if err != nil {
		log.Fatal(err)
		rollBackTx(itx.tx)
	}
	telapsed := time.Since(tstart)
	log.Printf("INFO: Done is %s", telapsed)
}
