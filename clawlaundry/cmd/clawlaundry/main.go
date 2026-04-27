// Command clawlaundry is the HTTP API for MiroClaw/ZeroClaw-style swarm agent metadata
// (Hands + cron jobs + defaults) backed by PostgreSQL and GORM, plus a prompt workspace CLI.
package main

import (
	"fmt"
	"log/slog"
	"os"
)

func main() {
	if err := newRoot().Execute(); err != nil {
		slog.Error("clawlaundry", "err", err)
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
