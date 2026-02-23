package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"
)

var (
	user     string
	url      string
	file     string
	interval time.Duration
	once     bool
	verbose  bool
	client   = http.DefaultClient
)

func main() {
	flag.StringVar(&user, "user", "", "Username to fetch keys for")
	flag.StringVar(&url, "url", "https://github.com", "Base URL to fetch keys from")
	flag.StringVar(&file, "file", "~/.ssh/authorized_keys", "File to update with keys (replaces content)")
	flag.DurationVar(&interval, "interval", time.Minute, "update interval (ex. 5s, 1m, 3h)")
	flag.BoolVar(&once, "once", false, "Run once and then exit")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.Parse()

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	if user == "" {
		slog.Error("missing -user")
		os.Exit(1)
	}

	if file[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			slog.Error("Could not determine home directory", slog.Any("error", err))
			os.Exit(1)
		}
		file = filepath.Join(home, file[1:])
	}

	if err := run(context.Background()); err != nil {
		slog.Error("Could not run ssh-key-sync", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	slog.Debug("Running initial sync")
	if err := sync(ctx); err != nil {
		return fmt.Errorf("could run initial sync: %w", err)
	}

	if once {
		return nil
	}

	ticker := time.NewTicker(interval)
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	slog.Debug("Starting ticker", slog.Duration("interval", interval))
	for {
		select {
		case <-ctx.Done():
			slog.Info("Stopping sync")
			return nil
		case <-ticker.C:
			slog.Debug("Running sync")
			if err := sync(ctx); err != nil {
				slog.Error("Could not sync", slog.Any("error", err))
			}
		}
	}
}

func sync(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url+"/"+user+".keys", nil)
	if err != nil {
		return fmt.Errorf("could not build request: %w", err)
	}

	slog.Debug("Fetching keys", slog.String("url", req.URL.String()))
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not fetch keys: %w", err)
	}
	defer resp.Body.Close()

	keys, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read keys: %w", err)
	}

	slog.Debug("Updating key file", slog.String("file", file))
	if err := os.WriteFile(file, keys, 0600); err != nil {
		return fmt.Errorf("could not write file: %w", err)
	}
	return nil
}
