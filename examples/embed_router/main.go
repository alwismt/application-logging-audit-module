// Embed the module HTTP routes in an existing Chi application.
//
// Run:
//
//	go run ./examples/embed_router/main.go
//
// Then open http://localhost:9090/health and http://localhost:9090/logs/log-info (POST).
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/alwismt/application-logging-audit-module/pkg/loggingaudit"
	"github.com/go-chi/chi/v5"
)

func main() {
	mod, err := loggingaudit.NewFromEnv()
	if err != nil {
		log.Fatalf("init module: %v", err)
	}

	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, "My app — logging/audit mounted below")
	})
	r.Mount("/", mod.Handler())

	// Host port for this demo app (module config still uses APP_PORT / .env for DB paths).
	addr := ":9090"
	if p := os.Getenv("EMBED_PORT"); p != "" {
		addr = ":" + p
	}
	fmt.Printf("Listening on %s\n", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}
