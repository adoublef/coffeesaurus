package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/adoublef-go/template"
	"github.com/go-chi/chi/v5"
)

//go:embed all:*.html
var fsys embed.FS

// change API to allow any pattern
var t = template.Must(fsys, template.Partials(false))

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

	mux := chi.NewMux()
	{
		// simple index page
		mux.Get("/", func(w http.ResponseWriter, r *http.Request) {
			// NOTE static fonts and styles handled by external project
			t.ExecuteHTTP(w, r, "index", nil)
		})
	}

	s := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
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
