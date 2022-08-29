package metric_stdprint

import (
	"context"
	"log"
	"math/rand"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
)

func newController(ctx context.Context) func()  {
	exporter, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
	if nil!=err{
		panic(err)
	}
	pusher := controller.New(processor.NewFactory(simple.NewWithInexpensiveDistribution(), exporter), controller.WithExporter(exporter))
	err = pusher.Start(ctx)
	if nil!=err{
		log.Fatalf("starting push controller: %v",err)
	}
	global.SetMeterProvider(pusher)
	return func() {
		if err := pusher.Stop(ctx);nil!=err{
			log.Fatalf("stopping push controller: %v",err)
		}
	}

}

func hao() {
	ctx := context.TODO()
	cleanup := newController(ctx)
	defer cleanup()

	/*	metricClient := otlpmetricgrpc.NewClient(
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint("1.1.1.1"))

	exporter, err := otlpmetric.New(ctx, metricClient)
	if nil!=err{
		panic(err)
	}
	pusher := controller.New(processor.NewFactory(simple.NewWithHistogramDistribution(), exporter), controller.WithExporter(exporter), controller.WithCollectPeriod(time.Second*2))
	global.SetMeterProvider(pusher)
	err = pusher.Start(ctx)
	if nil!=err{
		panic(err)
	}*/

	meter := global.Meter("demo-client-meter")

	commonLabels := []attribute.KeyValue{
		attribute.String("method","repl"),
		attribute.String("client","cli"),
	}


	requestLatency := metric.Must(meter).NewFloat64Histogram("demo_client/request_latency", metric.WithDescription("The latency of requests processed"))

	requestCount := metric.Must(meter).NewInt64Counter("demo_client/request_counts", metric.WithDescription("The number of requests preocessed"))

	lineLengths := metric.Must(meter).NewInt64Histogram("demo_client/line_lengths", metric.WithDescription("The lengths of the various line in"))

	lineCounts := metric.Must(meter).NewInt64Counter("demo_client/line_counts", metric.WithDescription("The counts of the lines in"))

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for {
		startTime := time.Now()
		latencyMs := float64(time.Since(startTime)) / 1e6
		nr := int(rng.Int31n(7))
		for i:=0;i<nr;i++{
			randLineLength := rng.Int63n(999)
			meter.RecordBatch(ctx,commonLabels,lineCounts.Measurement(1),lineLengths.Measurement(randLineLength))
			//fmt.Printf("#%d: LineLength: %dBy\n", i, randLineLength)
		}

		meter.RecordBatch(ctx,commonLabels,requestLatency.Measurement(latencyMs),requestCount.Measurement(1))
		//fmt.Printf("Latency: %.3fms\n",latencyMs)
		time.Sleep(time.Second)
	}
}
