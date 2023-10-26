package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/adoublef/coffeesaurus/env"
	iamHTTP "github.com/adoublef/coffeesaurus/internal/iam/http"
	"github.com/adoublef/coffeesaurus/internal/iam/sqlite3"
	"github.com/go-chi/chi/v5"
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
	// ping database
	db, err := sql.Open("sqlite3", os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("opening connection: %w", err)
	}
	defer db.Close()
	err = sqlite3.Ping(ctx, db, "profiles")
	if err != nil {
		return err
	}

	mux := chi.NewMux()
	{
		mux.Mount("/", iamHTTP.NewService())
	}

	s := &http.Server{
		Addr:    ":" + env.WithValue("PORT", "8080"),
		Handler: mux,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	sErr := make(chan error)
	go func() {
		sErr <- s.ListenAndServe()
	}()

	select {
	case err := <-sErr:
		return fmt.Errorf("main error: starting server: %w", err)
	case <-ctx.Done():
		// TODO
		return s.Shutdown(context.Background())
	}
}
