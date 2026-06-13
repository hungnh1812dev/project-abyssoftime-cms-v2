package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	deliveryhandler "project-abyssoftime-cms-v2/api/internal/delivery/http/handler"
	"project-abyssoftime-cms-v2/api/internal/delivery/http/middleware"
	"project-abyssoftime-cms-v2/api/internal/infrastructure/mongodb"
	"project-abyssoftime-cms-v2/api/internal/usecase/auth"
	contenttype "project-abyssoftime-cms-v2/api/internal/usecase/content_type"
	docuc "project-abyssoftime-cms-v2/api/internal/usecase/document"
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

	db := mongodb.Database(mongoClient)

	// repositories
	userRepo := mongodb.NewUserRepository(db)
	ctRepo := mongodb.NewContentTypeRepository(db)
	docRepo := mongodb.NewDocumentRepository(db)
	mediaRepo := mongodb.NewMediaAssetRepository(db)

	// usecases
	authUC := auth.New(userRepo)
	ctUC := contenttype.New(ctRepo)
	documentUC := docuc.New(docRepo, mediaRepo)

	// handlers
	authHandler := deliveryhandler.NewAuthHandler(authUC)
	ctHandler := deliveryhandler.NewContentTypeHandler(ctUC)
	docHandler := deliveryhandler.NewDocumentHandler(documentUC)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/refresh", authHandler.Refresh)
	mux.HandleFunc("POST /auth/logout", authHandler.Logout)

	adminOnly := func(h http.HandlerFunc) http.Handler {
		return middleware.Auth(middleware.RequireRole("admin", h))
	}
	mux.Handle("GET /api/content-types", adminOnly(ctHandler.List))
	mux.Handle("POST /api/content-types", adminOnly(ctHandler.Create))
	mux.Handle("GET /api/content-types/{id}", adminOnly(ctHandler.GetByID))
	mux.Handle("PUT /api/content-types/{id}", adminOnly(ctHandler.Update))
	mux.Handle("DELETE /api/content-types/{id}", adminOnly(ctHandler.Delete))

	authRequired := func(h http.HandlerFunc) http.Handler {
		return middleware.Auth(http.HandlerFunc(h))
	}
	mux.Handle("GET /api/documents", authRequired(docHandler.List))
	mux.Handle("GET /api/documents/{id}", authRequired(docHandler.GetByID))
	mux.Handle("POST /api/documents", adminOnly(docHandler.Create))
	mux.Handle("PUT /api/documents/{id}", adminOnly(docHandler.Update))
	mux.Handle("DELETE /api/documents/{id}", adminOnly(docHandler.Delete))
	mux.Handle("POST /api/documents/{id}/publish", adminOnly(docHandler.Publish))
	mux.Handle("POST /api/documents/{id}/unpublish", adminOnly(docHandler.Unpublish))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
