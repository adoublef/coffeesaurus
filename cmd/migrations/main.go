// migration script used for
//
// See more https://github.com/maragudk/litefs-app/blob/main/cmd/migrate/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	s3 "github.com/adoublef/coffeesaurus/internal/iam/sqlite3"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	q := make(chan os.Signal, 1)
	signal.Notify(q, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-q
		cancel()
	}()

	if err := run(ctx); err != nil {
		log.Printf("adoublef/coffeesaurus: %s", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) (err error) {
	// migrations for `iam` module
	err = s3.Up(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("running migration: %w", err)
	}
	return nil
}
