package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/panos-zamos/a-apps/apps/projects/handlers"
	"github.com/panos-zamos/a-apps/shared/auth"
	"github.com/panos-zamos/a-apps/shared/database"
	"github.com/panos-zamos/a-apps/shared/models"
	sharedTemplates "github.com/panos-zamos/a-apps/shared/templates"
)

func main() {
	// Load configuration
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-in-production"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3002"
	}

	// Open database
	db, err := database.Open("data/projects.db")
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
		users = []models.UserFromConfig{}
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

		// Project management
		r.Get("/projects/new", h.NewProjectForm)
		r.Post("/projects", h.CreateProject)
		r.Get("/projects/{id}", h.ProjectDetail)
		r.Get("/projects/{id}/edit", h.EditProjectForm)
		r.Put("/projects/{id}", h.UpdateProject)
		r.Put("/projects/{id}/stage", h.UpdateProjectStage)
		r.Delete("/projects/{id}", h.DeleteProject)

		// Log entries
		r.Post("/projects/{id}/log", h.CreateLogEntry)
		r.Get("/projects/{id}/log/{logId}/reply", h.ReplyForm)
		r.Post("/projects/{id}/log/{logId}/reply", h.CreateLogReply)
		r.Delete("/projects/{id}/log/{logId}", h.DeleteLogEntry)
	})

	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("projects starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
