package main //32twOqGf8gIswTgzG3IKxP

import (
	"database/sql"
	"fmt"
	"math"
	"net/http"
	"os"
	"runtime/pprof"
	"sync"

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

	// Create a CPU profile file.
	f, err := os.Create("beans.prof")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Start CPU profiling.
	if err := pprof.StartCPUProfile(f); err != nil {
		panic(err)
	}
	defer pprof.StopCPUProfile()

	cfp := os.Getenv("CONFIG_FILE")
	if cfp == "" {
		cfp = "config.yaml"
	}
	cfg := parseConfig(cfp)

	db, err := InitDB(cfg)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	c := &http.Client{}
	token := fetchAccessToken(c, cfg.Spotify.TokenEndpoint, cfg.Spotify.ApiKey).AccessToken
	filterStr := "fields=next,items(added_by.id,added_at,track(name,id,duration_ms,album(name),artists(name))"
	pageBaseUrl := fmt.Sprintf("%s/%s/tracks", cfg.Spotify.PlaylistEndpoint, cfg.Spotify.PlaylistId)
	metadataUrl := fmt.Sprintf("%s/%s?fields=snapshot_id,tracks.total", cfg.Spotify.PlaylistEndpoint, cfg.Spotify.PlaylistId)
	//nextPage := fmt.Sprintf("%s?%s&limit=100", pageBaseUrl, filterStr)

	// itx, err := StartTransaction(db)
	// if err != nil {
	// 	panic(err)
	// }

	// check if the playlist has changed, return early if there is nothing to import
	newSnapshotId, songCount := fetchMetadata(c, metadataUrl, token)
	// var oldSnapshotId string
	// if err := itx.statements.GetSnapshotId.QueryRow().Scan(&oldSnapshotId); err != nil {
	// 	panic(err)
	// }
	// if newSnapshotId == oldSnapshotId {
	// 	fmt.Println("Snapshot IDs match, nothing to import")
	// 	return
	// }

	// fetch data from spotify API concurrently using goroutines
	// calculate number of requests needed from playlist song count
	//
	// API rate limit ~180 requests per second, returns max 100 tracks per request
	// playlist max song count is 10,000
	// -> should be possible to always fetch all songs concurrently before rate limit kicks in
	// @TODO: add backoff, just in case :P
	fmt.Printf("songCount: %d\n", songCount)
	reqCount := int(math.Ceil((float64(songCount) / 100)))
	fmt.Printf("reqCount: %d\n", reqCount)
	tracksCh := make(chan []Track, reqCount)
	var wg sync.WaitGroup
	for i := 0; i < reqCount; i++ {
		wg.Add(1)
		apiUrl := fmt.Sprintf(
			"%s?%s&limit=100&offset=%d",
			pageBaseUrl, filterStr, i*100,
		)
		go func(apiUrl string) {
			fmt.Printf("fetching: %s\n", apiUrl)
			defer wg.Done()
			tracks := fetchPageData(c, apiUrl, token, newSnapshotId)
			fmt.Printf("goroutine fetched %d tracks\n", len(tracks))
			tracksCh <- tracks
		}(apiUrl)
	}
	wg.Wait()
	fmt.Println("All goroutines finished, closing channel")
	close(tracksCh)
	fmt.Print(len(tracksCh))

	// // update snapshot_id
	// _, err = itx.statements.UpdateSnapshotId.Exec(newSnapshotId)
	// if err != nil {
	// 	fmt.Println("unable to update snapshot id")
	// 	panic(err)
	// }

	// // Add new tracks, update snapshot_id where applicable
	// // @TODO: implement batching
	// for nextPage != "" {
	// 	pg := fetchPageData(c, nextPage, token)
	// 	for _, e := range pg.Items {
	// 		date, _ := time.Parse(time.RFC3339, e.AddedAt)
	// 		_, err := itx.statements.InsertUpdateTrack.Exec(
	// 			e.Track.ID,
	// 			e.Track.Name,
	// 			e.Track.Album.Name,
	// 			e.Track.Artists[0].Name,
	// 			e.AddedBy.ID,
	// 			date,
	// 			e.Track.DurationMs,
	// 			newSnapshotId,
	// 		)
	// 		if err != nil {
	// 			fmt.Printf("Failed to insert new track: %s \n", err)
	// 			rollBackTx(itx.tx)
	// 		}
	// 	}
	// 	fmt.Print(pg.Next)
	// 	nextPage = pg.Next
	// }

	// // delete old tracks
	// _, err = itx.statements.DeleteTracks.Exec(oldSnapshotId)
	// if err != nil {
	// 	fmt.Printf("Failed to delete old tracks: %s \n", err)
	// 	rollBackTx(itx.tx)
	// }

	// err = itx.tx.Commit()
	// if err != nil {
	// 	rollBackTx(itx.tx)
	// }
	// fmt.Println("Done!")
}
