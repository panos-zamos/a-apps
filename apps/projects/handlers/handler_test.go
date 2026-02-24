package handlers

import (
	"context"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/panos-zamos/a-apps/shared/auth"
	"github.com/panos-zamos/a-apps/shared/database"
)

func setupTestDB(t *testing.T) *database.DB {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "projects.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.RunMigrations(Migrations); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	return db
}

func TestHomeRendersProjects(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	username := "panos"
	projectName := "Test Project"

	_, err := db.Exec(
		`INSERT INTO projects (username, short_name, short_description, stage, rating) VALUES (?, ?, ?, ?, ?)`,
		username, projectName, "A test project", "idea", 3,
	)
	if err != nil {
		t.Fatalf("insert project: %v", err)
	}

	h := &Handler{DB: db, Users: nil, JWTSecret: "test"}

	projects, err := h.getProjects(username, "", "", 0)
	if err != nil {
		t.Fatalf("get projects: %v", err)
	}
	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}

	req := httptest.NewRequest("GET", "/", nil)
	ctx := context.WithValue(req.Context(), auth.UserContextKey, username)
	req = req.WithContext(ctx)
	resp := httptest.NewRecorder()

	h.Home(resp, req)

	if resp.Code != 200 {
		t.Fatalf("expected status 200, got %d", resp.Code)
	}
	body := resp.Body.String()
	if !strings.Contains(body, projectName) {
		t.Fatalf("expected project name in response")
	}
}

func TestFilterByStage(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	username := "panos"
	db.Exec(`INSERT INTO projects (username, short_name, stage) VALUES (?, ?, ?)`, username, "Proj A", "idea")
	db.Exec(`INSERT INTO projects (username, short_name, stage) VALUES (?, ?, ?)`, username, "Proj B", "released")

	h := &Handler{DB: db, Users: nil, JWTSecret: "test"}

	projects, _ := h.getProjects(username, "idea", "", 0)
	if len(projects) != 1 {
		t.Fatalf("expected 1 project with stage=idea, got %d", len(projects))
	}
	if projects[0].ShortName != "Proj A" {
		t.Fatalf("expected Proj A, got %s", projects[0].ShortName)
	}
}

func TestLogEntries(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	username := "panos"
	result, err := db.Exec(
		`INSERT INTO projects (username, short_name, stage) VALUES (?, ?, ?)`,
		username, "Test Project", "idea",
	)
	if err != nil {
		t.Fatalf("insert project: %v", err)
	}
	projectID, _ := result.LastInsertId()

	rootResult, err := db.Exec(
		`INSERT INTO log_entries (project_id, username, note, url) VALUES (?, ?, ?, ?)`,
		projectID, username, "Started research", "https://example.com",
	)
	if err != nil {
		t.Fatalf("insert log entry: %v", err)
	}
	rootID, _ := rootResult.LastInsertId()

	_, err = db.Exec(
		`INSERT INTO log_entries (project_id, parent_id, username, note) VALUES (?, ?, ?, ?)`,
		projectID, rootID, username, "Found useful resource",
	)
	if err != nil {
		t.Fatalf("insert child log entry: %v", err)
	}

	h := &Handler{DB: db, Users: nil, JWTSecret: "test"}

	entries, err := h.getLogEntries(int(projectID), username)
	if err != nil {
		t.Fatalf("get log entries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 root entry, got %d", len(entries))
	}
	if entries[0].Note != "Started research" {
		t.Fatalf("expected root note, got %q", entries[0].Note)
	}
	if len(entries[0].Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(entries[0].Children))
	}
	if entries[0].Children[0].Note != "Found useful resource" {
		t.Fatalf("expected child note, got %q", entries[0].Children[0].Note)
	}
}
