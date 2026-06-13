package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"project-abyssoftime-cms-v2/api/internal/infrastructure/mongodb"
)

func main() {
	ctx := context.Background()

	mongoClient, err := mongodb.NewClient(ctx)
	if err != nil {
		log.Fatalf("mongodb connect: %v", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("mongodb disconnect: %v", err)
		}
	}()

	log.Println("connected to mongodb")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	addr := ":" + port
	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
