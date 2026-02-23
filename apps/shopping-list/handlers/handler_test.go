package handlers

import (
	"context"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/panos/a-apps/shared/auth"
	"github.com/panos/a-apps/shared/database"
)

func TestHomeRendersSavedItems(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "shopping-list.db")

	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := db.RunMigrations(Migrations); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	username := "panos"
	storeName := "Test Store"
	itemName := "Milk"
	quantity := "1"

	storeResult, err := db.Exec(`
		INSERT INTO stores (name, username, color)
		VALUES (?, ?, ?)
	`, storeName, username, "#000000")
	if err != nil {
		t.Fatalf("insert store: %v", err)
	}

	storeID, err := storeResult.LastInsertId()
	if err != nil {
		t.Fatalf("store id: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO items (store_id, name, quantity, username, checked)
		VALUES (?, ?, ?, ?, ?)
	`, storeID, itemName, quantity, username, 0)
	if err != nil {
		t.Fatalf("insert item: %v", err)
	}

	h := &Handler{DB: db, Users: nil, JWTSecret: "test"}

	items, err := h.getItems(int(storeID), username)
	if err != nil {
		t.Fatalf("get items: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item from getItems, got %d", len(items))
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
	if !strings.Contains(body, storeName) {
		t.Fatalf("expected store name in response")
	}
	if !strings.Contains(body, itemName) {
		t.Fatalf("expected item name in response, body: %s", body)
	}
}
