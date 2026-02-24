package handlers

import (
	"html/template"
	"net/http"

	"github.com/panos-zamos/a-apps/shared/auth"
	"github.com/panos-zamos/a-apps/shared/database"
	"github.com/panos-zamos/a-apps/shared/models"
	sharedTemplates "github.com/panos-zamos/a-apps/shared/templates"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	DB        *database.DB
	Users     []models.UserFromConfig
	JWTSecret string
}

// Project represents a tracked project
type Project struct {
	ID               int
	Username         string
	ShortName        string
	ShortDescription string
	FullDescription  string
	WebsiteURL       string
	SourceURL        string
	IsCommercial     bool
	IsOpenSource     bool
	IsPublic         bool
	Stage            string
	Rating           int
	CreatedAt        string
	UpdatedAt        string
}

// LogEntry represents a timeline/research entry
type LogEntry struct {
	ID        int
	ProjectID int
	ParentID  *int
	Username  string
	Note      string
	URL       string
	CreatedAt string
	Children  []LogEntry
}

// Stages defines the valid project stages in order
var Stages = []string{"idea", "planning", "development", "released", "archived"}

// LoginPage renders the login page
func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.New("login").Parse(sharedTemplates.LoginHTML))
	data := map[string]interface{}{
		"AppName": "projects",
		"Error":   r.URL.Query().Get("error"),
	}
	tmpl.Execute(w, data)
}

// Login handles login form submission
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")

	valid, err := auth.ValidateCredentials(username, password, h.Users)
	if !valid || err != nil {
		http.Redirect(w, r, "/login?error=Invalid credentials", http.StatusSeeOther)
		return
	}

	token, err := auth.GenerateToken(username, h.JWTSecret)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Logout handles logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "auth_token",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// HealthCheck returns OK for health checks
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
