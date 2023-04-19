package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Listen
func listen(ctx context.Context, cancel context.CancelFunc, address string) func() error {
	return func() error {
		defer cancel()
		http.Handle("/metrics", promhttp.Handler())
		server := &http.Server{Addr: address}
		go func() {
			log.Printf("listener started on %s\n", address)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Println("listener error:", err)
			}
		}()
		<-ctx.Done()
		err := server.Shutdown(context.Background())
		if err != nil {
			log.Println("listener shutdown error:", err)
		}
		return err
	}
}

func controlC(ctx context.Context, cancel context.CancelFunc) func() error {
	return func() error {
		defer cancel()
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		for {
			select {
			case <-c:
				fmt.Println("\nCtrl+C stopped!")
				return nil
			case <-ctx.Done():
				return nil
			}
		}
	}
}
