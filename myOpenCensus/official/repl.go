package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"contrib.go.opencensus.io/exporter/prometheus"
)

//Create Metrics
var (
	//The latency in milliseconds
	MLatencyMs = stats.Float64("repl/latency", "The latency in milliseconds per REPL loop", "ms")

	//Counts/groups the lengths of lines read in.
	MLineLengths = stats.Int64("repl/line_lengths", "The distribution of line lengths", "By")
)

//Create Tags	简简单单加了一个标签而已,需要提前在view里注册,配合ctx使用
var (
	KeyMethod, _ = tag.NewKey("method")
	KeyStatus, _ = tag.NewKey("status")
	KeyError, _  = tag.NewKey("error")
)

//Create Views
var (
	LatencyView = &view.View{			//延迟, histogram
		Name:        "demo/latency",
		Measure:     MLatencyMs,
		Description: "The distribution of the latncies",

		//Latency in buckets:
		// [>=0ms, >=25ms, >=50ms, >=75ms, >=100ms, >=200ms, >=400ms, >=600ms, >=800ms, >=1s, >=2s, >=4s, >=6s]
		Aggregation: view.Distribution(0, 25, 50, 75, 100, 200, 400, 600, 800, 1000, 2000, 4000, 6000),
		TagKeys:     []tag.Key{KeyMethod}}		//latency注册了这个KeyMethod的tag

	LineCountView = &view.View{			//终端输入的次数, count
		Name:        "demo/lines_in",
		Measure:     MLineLengths,
		Description: "The number of lines from standard input",
		Aggregation: view.Count(),
		TagKeys: []tag.Key{KeyStatus},		//lineCount注册了KeyStatus的tag
	}

	LineLengthView = &view.View{	// histogram, 其下层有count,sum和bucket数据的记录
		Name:        "demo/line_lengths",
		Description: "Groups the lengths of keys in buckets",
		Measure:     MLineLengths,
		// Lengths: [>=0B, >=5B, >=10B, >=15B, >=20B, >=40B, >=60B, >=80, >=100B, >=200B, >=400, >=600, >=800, >=1000]
		Aggregation: view.Distribution(0, 5, 10, 15, 20, 40, 60, 80, 100, 200, 400, 600, 800, 1000),
	}
)

func main() {
	// Register The Views, it is imperative that this step exists
	// lest recorded metrics will be dropped and never exported.
	if err := view.Register(LatencyView, LineCountView, LineLengthView); nil != err {
		log.Fatalf("Failed to register views: %v", err)
	}

	// Create the Prometheus exporter
	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: "ocmetricstutorial",
	})
	if nil != err {
		log.Fatalf("Failed to create the Promethus stats exporter: %v", err)
	}

	//Now finally run the Prometheus exporter as a scrape endpoint.
	//We'll run the server on port 8888.
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", pe)
		if err := http.ListenAndServe(":8888", mux); err != nil {
			log.Fatalf("Failed to run Prometheus scrape endpoint: %v", err)
		}
	}()

	//In a REPL:
	//	1. Read input
	//  2. process input
	br := bufio.NewReader(os.Stdin)

	// Register the views
	if err := view.Register(LatencyView, LineCountView, LineLengthView); err != nil {
		log.Fatalf("Failed to register views: %v", err)
	}

	//repl is the read, evaluate, print, loop
	for {
		if err := readEvaluateProcess(br); nil != err {
			if err == io.EOF {
				return
			}
			log.Fatal(err)
		}
	}
}

// readEvaluateProcess reads a line from the input reader and
// then processes it. It returns an error if any was encountered.
func readEvaluateProcess(br *bufio.Reader) (terr error) {
	startTime := time.Now()
	//Inserting Tags
	ctx, err := tag.New(context.Background(), tag.Insert(KeyMethod, "repl"), tag.Insert(KeyStatus, "OK"))
	if nil != err {
		return err
	}

	defer func() {
		if terr != nil {
			ctx, _ = tag.New(ctx, tag.Upsert(KeyStatus, "ERROR"),
				tag.Upsert(KeyError, terr.Error()))
		}

		stats.Record(ctx, MLatencyMs.M(sinceInMilliseconds(startTime)))
	}()

	fmt.Printf("> ")
	line, _, err := br.ReadLine()
	if nil != err {
		if err != io.EOF{
			return err
		}
		log.Fatal(err)
	}

	out, err := processLine(ctx, line)
	if nil != err {
		return err
	}
	fmt.Printf("< %s\n\n", out)
	return nil
}

// processLine takes in a line of text and
// transforms it. Currently it just capitalizes it.
func processLine(ctx context.Context, in []byte) (out []byte, err error) {
	startTime := time.Now()
	defer func() {
		stats.Record(ctx, MLatencyMs.M(sinceInMilliseconds(startTime)),
			MLineLengths.M(int64(len(in))))
	}()

	return bytes.ToUpper(in), nil
}

func sinceInMilliseconds(startTime time.Time) float64 {
	return float64(time.Since(startTime).Nanoseconds()) / 1e6
}
