package handlers

// Migrations contains SQL migration scripts
var Migrations = []string{
	`CREATE TABLE IF NOT EXISTS projects (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		short_name TEXT NOT NULL,
		short_description TEXT DEFAULT '',
		full_description TEXT DEFAULT '',
		website_url TEXT DEFAULT '',
		source_url TEXT DEFAULT '',
		is_commercial BOOLEAN DEFAULT 0,
		is_open_source BOOLEAN DEFAULT 0,
		is_public BOOLEAN DEFAULT 0,
		stage TEXT DEFAULT 'idea',
		rating INTEGER DEFAULT 0,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS log_entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_id INTEGER NOT NULL,
		parent_id INTEGER,
		username TEXT NOT NULL,
		note TEXT NOT NULL,
		url TEXT DEFAULT '',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
		FOREIGN KEY (parent_id) REFERENCES log_entries(id) ON DELETE CASCADE
	)`,
	`CREATE INDEX IF NOT EXISTS idx_projects_username ON projects(username)`,
	`CREATE INDEX IF NOT EXISTS idx_log_entries_project ON log_entries(project_id)`,
	`CREATE INDEX IF NOT EXISTS idx_log_entries_parent ON log_entries(parent_id)`,
	`CREATE INDEX IF NOT EXISTS idx_log_entries_username ON log_entries(username)`,
}
