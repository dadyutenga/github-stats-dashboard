package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github-stats-dashboard/api"
	"github-stats-dashboard/config"
	"github-stats-dashboard/renderer"
)

func main() {
	cfg := config.Load()
	if cfg.Token == "" {
		fmt.Println("Error: GITHUB_TOKEN environment variable not set.")
		fmt.Println("Run: export GITHUB_TOKEN=your_personal_access_token")
		os.Exit(1)
	}

	client := api.NewClient(cfg.Token, cfg.Username)

	// Handle Ctrl+C gracefully
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		renderer.ClearScreen()
		renderer.ShowCursor()
		fmt.Println("\n👋 Dashboard closed. Later!")
		os.Exit(0)
	}()

	renderer.HideCursor()

	for {
		stats, err := client.FetchStats()
		if err != nil {
			renderer.RenderError(err)
		} else {
			renderer.Render(stats)
		}

		// Wait custom seconds but allow early exit
		select {
		case <-time.After(time.Duration(cfg.RefreshSeconds) * time.Second):
		}
	}
}
