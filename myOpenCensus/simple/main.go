package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"time"

	"contrib.go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
)
// Create measures. The program will record measures for the size of
// processed videos and the number of videos marked as spam.
var (
	videoCount = stats.Int64("example.com/measures/video_count", "number of processed videos", stats.UnitDimensionless)
	videoSize  = stats.Int64("example.com/measures/video_size", "size of processed video", stats.UnitBytes)
)

func main() {
	ctx := context.Background()

	exporter, err := prometheus.NewExporter(prometheus.Options{})
	if err != nil {
		log.Fatal(err)
	}

	// Create view to see the number of processed videos cumulatively.
	// Create view to see the amount of video processed
	// Subscribe will allow view data to be exported.
	// Once no longer needed, you can unsubscribe from the view.

	if err = view.Register(
		&view.View{
			Name:        "video_count",
			Description: "number of videos processed over time",
			Measure:     videoCount,
			Aggregation: view.Count(),
		},
		&view.View{
			Name:        "video_size",
			Description: "processed video size over time",
			Measure:     videoSize,
			Aggregation: view.Distribution(0, 1<<16, 1<<32),
		},
	); err != nil {
		log.Fatalf("Cannot register the view: %v", err)
	}

	// Record some data points...
	go func() {
		for {
			stats.Record(ctx, videoCount.M(1), videoSize.M(rand.Int63()))
			<-time.After(time.Millisecond * time.Duration(1+rand.Intn(400)))
		}
	}()

	addr := ":9999"
	log.Printf("Serving at %s", addr)
	http.Handle("/metrics", exporter)
	log.Fatal(http.ListenAndServe(addr, nil))
}
