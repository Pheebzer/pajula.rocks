package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"sync"
	"time"

	"pajula.rocks/internal/config"
	"pajula.rocks/internal/db"
	imp "pajula.rocks/internal/importer"
	logger "pajula.rocks/internal/log"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	tstart := time.Now()
	logger.InitLogs()
	cfg := config.GetConfigs()
	conn, err := db.InitDB(cfg.MysqlDsn)
	if err != nil {
		log.Fatalf("ERROR: Unable to connect to database:\n%s", err)
	}
	defer conn.Close()

	c := &http.Client{}
	token := imp.FetchAccessToken(c, cfg.TokenEndpoint, cfg.ApiKey).AccessToken
	filterStr := "fields=next,items(added_by.id,added_at,track(name,id,duration_ms,album(name),artists(name))"
	pageBaseUrl := fmt.Sprintf("%s/%s/tracks", cfg.PlaylistEndpoint, cfg.PlaylistId)
	metadataUrl := fmt.Sprintf("%s/%s?fields=snapshot_id,tracks.total", cfg.PlaylistEndpoint, cfg.PlaylistId)

	itx, err := db.StartTransaction(conn)
	if err != nil {
		log.Fatal(err)
	}

	// check if the playlist has changed, return early if there is nothing to import
	newSnapshotId, songCount := imp.FetchMetadata(c, metadataUrl, token)
	var oldSnapshotId string
	if err := itx.Statements.GetSnapshotId.QueryRow().Scan(&oldSnapshotId); err != nil {
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
	tracksCh := make(chan []imp.Track, reqCount)
	var wg sync.WaitGroup

	log.Printf("INFO: Fetching songs for playlist %s", cfg.PlaylistId)

	for i := 0; i < reqCount; i++ {
		wg.Add(1)
		apiUrl := fmt.Sprintf("%s?%s&limit=100&offset=%d", pageBaseUrl, filterStr, i*100)
		go func(apiUrl string) {
			defer wg.Done()
			tracks := imp.FetchPageData(c, apiUrl, token, newSnapshotId)
			tracksCh <- tracks
		}(apiUrl)
	}

	wg.Wait()
	close(tracksCh)

	// update snapshot_id
	_, err = itx.Statements.UpdateSnapshotId.Exec(newSnapshotId)
	if err != nil {
		log.Printf("ERROR: Failed to update snapshot ID:\n%s", err)
		abort(itx)
	}

	// Add new tracks, update snapshot_id where applicable
	for ts := range tracksCh {
		for _, t := range ts {
			_, err := itx.Statements.InsertUpdateTrack.Exec(
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
				log.Printf("ERROR: Failed to update database:\n%s", err)
				abort(itx)
			}
		}
	}

	// delete old tracks, commit changes
	if _, err = itx.Statements.DeleteTracks.Exec(oldSnapshotId); err != nil {
		log.Printf("ERROR: Failed to delete old tracks:\n%s", err)
		abort(itx)
	}
	if err = db.Commit(itx); err != nil {
		log.Printf("ERROR: Failed to commit transaction:\n%s", err)
		abort(itx)
	}

	telapsed := time.Since(tstart)
	log.Printf("INFO: Done is %s", telapsed)
}

// wrapper function for tx rollback
func abort(itx *db.ImporterTx) {
	if err := db.RollBackTx(itx); err != nil {
		log.Fatalf("ERROR: Unable to roll back transaction:\n%s", err)
	}
	os.Exit(1)
}
