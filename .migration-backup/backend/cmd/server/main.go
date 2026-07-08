package main

import (
	"log"
	"net/http"
	"os"

	"avex-backend/internal/admin"
	"avex-backend/internal/customer"
	"avex-backend/internal/driver"
	"avex-backend/internal/merchant"
	"avex-backend/internal/shared"
	"avex-backend/internal/support"

	"github.com/rs/cors"
)

func main() {
	if err := shared.InitDB(); err != nil {
		log.Fatalf("❌ DB: %v", err)
	}
	shared.Seed()

	mux := http.NewServeMux()

	// Register routes from all app packages
	customer.RegisterRoutes(mux)
	driver.RegisterRoutes(mux)
	merchant.RegisterRoutes(mux)
	support.RegisterRoutes(mux)
	admin.RegisterRoutes(mux)

	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type", "Accept"},
		AllowCredentials: false,
	}).Handler(mux)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 AVEX API on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
