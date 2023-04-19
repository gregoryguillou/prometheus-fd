package main

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/sync/errgroup"
)

func main() {
	config, err := Parse()
	if err != nil {
		fmt.Println("could not parse content, error:", err)
		os.Exit(2)
	}
	g, ctx := errgroup.WithContext(context.Background())
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	g.Go(listen(ctx, cancel, config.Listener))
	g.Go(controlC(ctx, cancel))
	g.Go(monitor(ctx, cancel, config))
	g.Wait()
}
