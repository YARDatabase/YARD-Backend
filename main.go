package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"yard-backend/internal/config"
	"yard-backend/internal/handlers"
	"yard-backend/internal/middleware"
	"yard-backend/internal/services"
)

// main entry point initializes config redis texture loader and starts the http server
func main() {
	config.LoadEnv()
	config.InitRedis()
	handlers.InitTextureLoader()
	
	if err := services.LoadNEUReforgeStones(); err != nil {
		log.Printf("Warning: Failed to load NEU reforge stones: %v", err)
	}
	
	if err := services.LoadNEUReforges(); err != nil {
		log.Printf("Warning: Failed to load NEU reforges: %v", err)
	}
	
	services.StartScheduler()

	r := mux.NewRouter()
	r.HandleFunc("/health", handlers.HandleHealth).Methods("GET")
	r.HandleFunc("/api/reforge-stones", middleware.RateLimitMiddleware(handlers.HandleReforgeStones)).Methods("GET")
	r.HandleFunc("/api/reforges", middleware.RateLimitMiddleware(handlers.HandleReforges)).Methods("GET")
	r.HandleFunc("/api/item/{itemId}", middleware.RateLimitMiddleware(handlers.HandleItemImage)).Methods("GET")
	r.HandleFunc("/api/item-data/{itemId}", middleware.RateLimitMiddleware(handlers.HandleItemImageByData)).Methods("GET")

	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("YARD Backend server starting on :8080")
	log.Fatal(server.ListenAndServe())
}
