package handlers

// Migrations contains SQL migration scripts
var Migrations = []string{
	`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS stores (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		username TEXT NOT NULL,
		color TEXT DEFAULT '#3B82F6',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(name, username)
	)`,
	`CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		store_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		quantity TEXT DEFAULT '',
		checked BOOLEAN DEFAULT 0,
		username TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (store_id) REFERENCES stores(id) ON DELETE CASCADE
	)`,
	`CREATE INDEX IF NOT EXISTS idx_items_store ON items(store_id)`,
	`CREATE INDEX IF NOT EXISTS idx_items_username ON items(username)`,
	`CREATE INDEX IF NOT EXISTS idx_stores_username ON stores(username)`,
}
