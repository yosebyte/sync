package main

import (
	"context"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/yosebyte/x/log"
)

var (
	logger  = log.NewLogger(log.Info, true)
	version = "dev"
	mu      sync.Mutex
)

func main() {
	parsedURL := getParsedURL(os.Args)
	query := parsedURL.Query()
	ivl := query.Get("ivl")
	src := query.Get("src")
	dst := query.Get("dst")
	if ivl == "" || src == "" || dst == "" {
		logger.Fatal("Missing required parameters")
		getExitInfo()
	}
	interval, err := strconv.Atoi(ivl)
	if err != nil {
		logger.Fatal("Invalid interval: %v", err)
		getExitInfo()
	}
	logger.Info("Syncing files: %v -> %v", src, dst)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	ticker := time.NewTicker(time.Duration(interval) * time.Hour)
	defer ticker.Stop()
	logger.Info("Starting initial sync...")
	runSync(ctx, src, dst)
	for {
		select {
		case <-ctx.Done():
			logger.Info("Shutting down...")
			mu.Lock()
			mu.Unlock()
			logger.Info("Shutdown complete")
			return
		case <-ticker.C:
			logger.Info("Starting scheduled sync...")
			runSync(ctx, src, dst)
		}
	}
}

func getParsedURL(args []string) *url.URL {
	if len(args) < 2 {
		getExitInfo()
	}
	parsedURL, err := url.Parse(args[1])
	if err != nil {
		logger.Fatal("Parse failed: %v", err)
		getExitInfo()
	}
	return parsedURL
}

func getExitInfo() {
	logger.SetLogLevel(log.Info)
	logger.Info(`Version: %v %v/%v

Usage:
    sync "cmd://<ivl=hour>&<src=dir>&<dst=dir>"

Example:
    sync "cmd://ivl=12&src=/path/to/source&dst=/path/to/target"
`, version, runtime.GOOS, runtime.GOARCH)
	os.Exit(1)
}
