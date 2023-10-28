package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/adoublef/coffeesaurus/env"
	iamHTTP "github.com/adoublef/coffeesaurus/internal/iam/http"
	"github.com/adoublef/coffeesaurus/internal/iam/sessions"
	"github.com/adoublef/coffeesaurus/sqlite3"
	"github.com/go-chi/chi/v5"
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
	// sessions
	ss, err := sessions.New(ctx, env.Must("DATABASE_URL_SESSIONS"))
	if err != nil {
		return fmt.Errorf("opening sessions db: %w", err)
	}
	defer ss.Close()

	// iam
	db, err := sqlite3.Open(env.Must("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("opening connection: %w", err)
	}
	defer db.Close()

	mux := chi.NewMux()
	{
		mux.Mount("/", iamHTTP.NewService(db, ss))
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
