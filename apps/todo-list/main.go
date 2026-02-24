package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/panos/a-apps/apps/todo-list/handlers"
	"github.com/panos/a-apps/shared/auth"
	"github.com/panos/a-apps/shared/database"
	"github.com/panos/a-apps/shared/models"
	sharedTemplates "github.com/panos/a-apps/shared/templates"
)

func main() {
	// Load configuration
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-in-production"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	// Open database
	db, err := database.Open("data/todo-list.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.RunMigrations(handlers.Migrations); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Load users from config
	users, err := auth.LoadUsersFromConfig("config.yaml")
	if err != nil {
		log.Printf("Warning: Failed to load users from config: %v", err)
		users = []models.UserFromConfig{} // Empty list, will need to add users
	}

	// Setup router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Create handler
	h := &handlers.Handler{
		DB:        db,
		Users:     users,
		JWTSecret: jwtSecret,
	}

	// Public routes
	r.Get("/login", h.LoginPage)
	r.Post("/login", h.Login)
	r.Post("/logout", h.Logout)
	r.Get("/health", h.HealthCheck)
	r.Get("/custom.css", sharedTemplates.CustomCSSHandler())

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(jwtSecret))
		r.Get("/", h.Home)

		// Store management
		r.Get("/stores/new", h.NewStoreForm)
		r.Post("/stores", h.CreateStore)
		r.Delete("/stores/{id}", h.DeleteStore)

		// Item management
		r.Post("/stores/{storeID}/items", h.CreateItem)
		r.Post("/items/{id}/toggle", h.ToggleItem)
		r.Delete("/items/{id}", h.DeleteItem)
	})

	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("todo-list starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
