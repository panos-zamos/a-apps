package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

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
	appConfig, err := models.LoadAppConfig("config.yaml")
	if err != nil {
		log.Printf("Warning: Failed to load config: %v", err)
		appConfig = models.AppConfig{}
	}
	if appConfig.ChangelogPath == "" {
		appConfig.ChangelogPath = "changelog.yaml"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = appConfig.JWTSecret
	}
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-in-production"
	}

	port := os.Getenv("PORT")
	if port == "" {
		if appConfig.Port != 0 {
			port = fmt.Sprintf("%d", appConfig.Port)
		} else {
			port = "3002"
		}
	}

	basePath := normalizeBasePath(os.Getenv("BASE_PATH"))

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
	users := appConfig.Users
	if len(users) == 0 {
		log.Printf("Warning: No users configured in config.yaml")
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
		AppConfig: appConfig,
		BasePath:  basePath,
	}

	// Public routes
	r.Get("/login", h.LoginPage)
	r.Post("/login", h.Login)
	r.Post("/logout", h.Logout)
	r.Get("/health", h.HealthCheck)
	r.Get("/changelog", h.ChangelogPage)
	r.Get("/custom.css", sharedTemplates.CustomCSSHandler())

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(auth.Middleware(jwtSecret, basePath))
		r.Get("/", h.Home)

		// Project management
		r.Get("/new", h.NewProjectForm)
		r.Post("/", h.CreateProject)
		r.Get("/{id}", h.ProjectDetail)
		r.Get("/{id}/edit", h.EditProjectForm)
		r.Put("/{id}", h.UpdateProject)
		r.Put("/{id}/stage", h.UpdateProjectStage)
		r.Delete("/{id}", h.DeleteProject)

		// Log entries
		r.Post("/{id}/log", h.CreateLogEntry)
		r.Get("/{id}/log/{logId}/reply", h.ReplyForm)
		r.Post("/{id}/log/{logId}/reply", h.CreateLogReply)
		r.Delete("/{id}/log/{logId}", h.DeleteLogEntry)
	})

	// Start server
	addr := fmt.Sprintf(":%s", port)
	log.Printf("projects starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func normalizeBasePath(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || value == "/" {
		return ""
	}
	if !strings.HasPrefix(value, "/") {
		value = "/" + value
	}
	return strings.TrimRight(value, "/")
}
