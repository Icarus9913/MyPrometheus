package metric_stdprint

import (
	"context"
	"log"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
)

const (
	instrumentationName    = "github.com/instrumentron"
	instrumentationVersion = "v0.1.0"
)

var (
	meter = global.GetMeterProvider().Meter(
		instrumentationName,
		metric.WithInstrumentationVersion(instrumentationVersion),
	)

	// counter
	loopCounter = metric.Must(meter).NewInt64Counter("function.loops",metric.WithDescription("test with counter"))

	// histogram
	paramValue  = metric.Must(meter).NewInt64Histogram("function.param", metric.WithDescription("test with histogram"))

	// label
	nameKey = attribute.Key("function.name")
)

func add(ctx context.Context, x, y int64) int64 {

	// label value
	nameKV := nameKey.String("add")


	loopCounter.Add(ctx, 1, nameKV)
	paramValue.Record(ctx, x, nameKV)
	paramValue.Record(ctx, y, nameKV)

	return x + y
}

func multiply(ctx context.Context, x, y int64) int64 {

	// label value
	nameKV := nameKey.String("multiply")

	loopCounter.Add(ctx, 1, nameKV)
	paramValue.Record(ctx, x, nameKV)
	paramValue.Record(ctx, y, nameKV)

	return x * y
}

func InstallExportPipeline(ctx context.Context) func() {
	exporter, err := stdoutmetric.New(stdoutmetric.WithPrettyPrint())
	if err != nil {
		log.Fatalf("creating stdoutmetric exporter: %v", err)
	}

	pusher := controller.New(
		// simple.NewWithInexpensiveDistribution这个selector最轻，使用的是minmaxsumcount
		// 也可以选用simple.NewWithHistogramDistribution()，一般的默认选择
		processor.NewFactory(simple.NewWithInexpensiveDistribution(),exporter),
		controller.WithExporter(exporter),
	)
	if err = pusher.Start(ctx); err != nil {
		log.Fatalf("starting push controller: %v", err)
	}
	global.SetMeterProvider(pusher)

	return func() {
		if err := pusher.Stop(ctx); err != nil {
			log.Fatalf("stopping push controller: %v", err)
		}
	}
}
