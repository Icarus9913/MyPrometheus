package metric_stdprint

import (
	"context"
	"log"
	"testing"
)

func TestDemo(t *testing.T)  {
	ctx := context.Background()

	// Registers a meter Provider globally.
	cleanup := InstallExportPipeline(ctx)
	defer cleanup()

	//log.Println("the answer is", add(ctx, multiply(ctx, multiply(ctx, 2, 2), 10), 2))
	log.Println("the answer is", add(ctx, 1,2))
	log.Println("the answer is", add(ctx, 3,4))
	log.Println("the answer is", multiply(ctx, 5,6))

}
