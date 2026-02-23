package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
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
	client   = http.DefaultClient
)

func main() {
	flag.StringVar(&user, "user", "", "Username to fetch keys for")
	flag.StringVar(&url, "url", "https://github.com", "Base URL to fetch keys from")
	flag.StringVar(&file, "file", "~/.ssh/authorized_keys", "File to update with keys (replaces content)")
	flag.BoolVar(&once, "once", false, "Run once and then exit")
	flag.DurationVar(&interval, "interval", time.Minute, "update interval (ex. 5s, 1m, 3h)")
	flag.Parse()

	if user == "" {
		log.Fatal("missing -user")
	}

	if file[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal("Could not determine home directory")
		}
		file = filepath.Join(home, file[1:])
	}

	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	if err := sync(ctx); err != nil {
		return fmt.Errorf("could sync: %w", err)
	}

	if once {
		return nil
	}

	ticker := time.NewTicker(interval)
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := sync(ctx); err != nil {
				log.Print(err.Error())
			}
		}
	}
}

func sync(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url+"/"+user+".keys", nil)
	if err != nil {
		return fmt.Errorf("could not build request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not fetch keys: %w", err)
	}
	defer resp.Body.Close()

	keys, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read keys: %w", err)
	}

	if err := os.WriteFile(file, keys, 0600); err != nil {
		return fmt.Errorf("could not write file: %w", err)
	}
	return nil
}
