package metric_grpc

import (
	"context"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/global"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
)

func tt() {
	ctx := context.Background()
	client := otlpmetricgrpc.NewClient(otlpmetricgrpc.WithInsecure(),otlpmetricgrpc.WithEndpoint("0.0.0.0"))
	exp, err := otlpmetric.New(ctx, client)
	if err != nil {
		log.Fatalf("Failed to create the collector exporter: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		if err := exp.Shutdown(ctx); err != nil {
			otel.Handle(err)
		}
	}()

	pusher := controller.New(
		processor.NewFactory(
			simple.NewWithHistogramDistribution(),
			exp,
		),
		controller.WithExporter(exp),
		controller.WithCollectPeriod(2*time.Second),
	)
	global.SetMeterProvider(pusher)

	if err := pusher.Start(ctx); err != nil {
		log.Fatalf("could not start metric controoler: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()
		// pushes any last exports to the receiver
		if err := pusher.Stop(ctx); err != nil {
			otel.Handle(err)
		}
	}()

	meter := global.Meter("test-meter")

	// Recorder metric example
	counter := metric.Must(meter).
		NewFloat64Counter(
			"an_important_metric",
			metric.WithDescription("Measures the cumulative epicness of the app"),
		)

	for i := 0; i < 10; i++ {
		log.Printf("Doing really hard work (%d / 10)\n", i+1)
		counter.Add(ctx, 1.0)
	}
}

