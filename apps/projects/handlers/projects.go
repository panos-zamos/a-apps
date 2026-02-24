package handlers

import (
	"database/sql"
	"fmt"
	"html"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/panos-zamos/a-apps/shared/auth"
	sharedTemplates "github.com/panos-zamos/a-apps/shared/templates"
)

// getProjects retrieves all projects for a user with optional filters
func (h *Handler) getProjects(username, stageFilter, typeFilter string, ratingFilter int) ([]Project, error) {
	query := `SELECT id, username, short_name, short_description, full_description,
		website_url, source_url, is_commercial, is_open_source, is_public,
		stage, rating, created_at, updated_at
		FROM projects WHERE username = ?`
	args := []interface{}{username}

	if stageFilter != "" {
		query += " AND stage = ?"
		args = append(args, stageFilter)
	}
	if typeFilter == "commercial" {
		query += " AND is_commercial = 1"
	} else if typeFilter == "open-source" {
		query += " AND is_open_source = 1"
	} else if typeFilter == "public" {
		query += " AND is_public = 1"
	}
	if ratingFilter > 0 {
		query += " AND rating = ?"
		args = append(args, ratingFilter)
	}

	query += " ORDER BY updated_at DESC"

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Username, &p.ShortName, &p.ShortDescription,
			&p.FullDescription, &p.WebsiteURL, &p.SourceURL,
			&p.IsCommercial, &p.IsOpenSource, &p.IsPublic,
			&p.Stage, &p.Rating, &p.CreatedAt, &p.UpdatedAt); err != nil {
			continue
		}
		projects = append(projects, p)
	}
	return projects, nil
}

// getProject retrieves a single project by ID
func (h *Handler) getProject(id int, username string) (*Project, error) {
	var p Project
	err := h.DB.QueryRow(`SELECT id, username, short_name, short_description, full_description,
		website_url, source_url, is_commercial, is_open_source, is_public,
		stage, rating, created_at, updated_at
		FROM projects WHERE id = ? AND username = ?`, id, username).Scan(
		&p.ID, &p.Username, &p.ShortName, &p.ShortDescription,
		&p.FullDescription, &p.WebsiteURL, &p.SourceURL,
		&p.IsCommercial, &p.IsOpenSource, &p.IsPublic,
		&p.Stage, &p.Rating, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// Home renders the projects listing page
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.GetUsername(r)

	stageFilter := r.URL.Query().Get("stage")
	typeFilter := r.URL.Query().Get("type")
	ratingFilter, _ := strconv.Atoi(r.URL.Query().Get("rating"))

	projects, _ := h.getProjects(username, stageFilter, typeFilter, ratingFilter)

	// Check if this is an HTMX request for just the table
	if r.Header.Get("HX-Request") == "true" {
		w.Write([]byte(h.projectsTable(projects)))
		return
	}

	data := map[string]interface{}{
		"Title":    "Projects",
		"AppName":  "Projects",
		"Username": username,
		"Content":  template.HTML(h.homeContent(projects, stageFilter, typeFilter, ratingFilter)),
	}

	tmpl := template.Must(template.New("base").Parse(sharedTemplates.BaseHTML))
	tmpl.Execute(w, data)
}

// NewProjectForm returns the form for creating a new project
func (h *Handler) NewProjectForm(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(h.projectForm(nil)))
}

// CreateProject handles creating a new project
func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.GetUsername(r)
	shortName := strings.TrimSpace(r.FormValue("short_name"))

	if shortName == "" {
		http.Error(w, "Project name is required", http.StatusBadRequest)
		return
	}

	isCommercial := r.FormValue("is_commercial") == "on"
	isOpenSource := r.FormValue("is_open_source") == "on"
	isPublic := r.FormValue("is_public") == "on"
	rating, _ := strconv.Atoi(r.FormValue("rating"))

	_, err := h.DB.Exec(`
		INSERT INTO projects (username, short_name, short_description, full_description,
			website_url, source_url, is_commercial, is_open_source, is_public, stage, rating)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		username, shortName,
		strings.TrimSpace(r.FormValue("short_description")),
		strings.TrimSpace(r.FormValue("full_description")),
		strings.TrimSpace(r.FormValue("website_url")),
		strings.TrimSpace(r.FormValue("source_url")),
		isCommercial, isOpenSource, isPublic,
		r.FormValue("stage"), rating)

	if err != nil {
		http.Error(w, "Failed to create project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/")
}

// EditProjectForm returns the edit form for a project
func (h *Handler) EditProjectForm(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.GetUsername(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	project, err := h.getProject(id, username)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	w.Write([]byte(h.projectForm(project)))
}

// UpdateProject handles updating a project
func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.GetUsername(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	shortName := strings.TrimSpace(r.FormValue("short_name"))
	if shortName == "" {
		http.Error(w, "Project name is required", http.StatusBadRequest)
		return
	}

	isCommercial := r.FormValue("is_commercial") == "on"
	isOpenSource := r.FormValue("is_open_source") == "on"
	isPublic := r.FormValue("is_public") == "on"
	rating, _ := strconv.Atoi(r.FormValue("rating"))

	_, err := h.DB.Exec(`
		UPDATE projects SET short_name = ?, short_description = ?, full_description = ?,
			website_url = ?, source_url = ?, is_commercial = ?, is_open_source = ?,
			is_public = ?, stage = ?, rating = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND username = ?`,
		shortName,
		strings.TrimSpace(r.FormValue("short_description")),
		strings.TrimSpace(r.FormValue("full_description")),
		strings.TrimSpace(r.FormValue("website_url")),
		strings.TrimSpace(r.FormValue("source_url")),
		isCommercial, isOpenSource, isPublic,
		r.FormValue("stage"), rating, id, username)

	if err != nil {
		http.Error(w, "Failed to update project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", fmt.Sprintf("/projects/%d", id))
}

// UpdateProjectStage handles inline stage change via HTMX
func (h *Handler) UpdateProjectStage(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.GetUsername(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))
	stage := r.FormValue("stage")

	_, err := h.DB.Exec(`
		UPDATE projects SET stage = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND username = ?`, stage, id, username)

	if err != nil {
		http.Error(w, "Failed to update stage", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteProject handles deleting a project
func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.GetUsername(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	_, err := h.DB.Exec(`DELETE FROM projects WHERE id = ? AND username = ?`, id, username)
	if err != nil {
		http.Error(w, "Failed to delete project", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/")
}

// ProjectDetail renders a single project detail page
func (h *Handler) ProjectDetail(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.GetUsername(r)
	id, _ := strconv.Atoi(chi.URLParam(r, "id"))

	project, err := h.getProject(id, username)
	if err != nil {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	logEntries, _ := h.getLogEntries(id, username)

	data := map[string]interface{}{
		"Title":    project.ShortName,
		"AppName":  "Projects",
		"Username": username,
		"Content":  template.HTML(h.detailContent(project, logEntries)),
	}

	tmpl := template.Must(template.New("base").Parse(sharedTemplates.BaseHTML))
	tmpl.Execute(w, data)
}

// --- Log Entry Handlers ---

// getLogEntries retrieves all log entries for a project, organized as a tree
func (h *Handler) getLogEntries(projectID int, username string) ([]LogEntry, error) {
	rows, err := h.DB.Query(`
		SELECT id, project_id, parent_id, username, note, url, created_at
		FROM log_entries
		WHERE project_id = ? AND username = ?
		ORDER BY created_at DESC`, projectID, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []LogEntry
	for rows.Next() {
		var e LogEntry
		var parentID sql.NullInt64
		if err := rows.Scan(&e.ID, &e.ProjectID, &parentID, &e.Username, &e.Note, &e.URL, &e.CreatedAt); err != nil {
			continue
		}
		if parentID.Valid {
			pid := int(parentID.Int64)
			e.ParentID = &pid
		}
		all = append(all, e)
	}

	// Build tree: separate roots and children
	childrenMap := make(map[int][]LogEntry)
	var roots []LogEntry
	for _, e := range all {
		if e.ParentID == nil {
			roots = append(roots, e)
		} else {
			childrenMap[*e.ParentID] = append(childrenMap[*e.ParentID], e)
		}
	}

	// Attach children to roots
	for i := range roots {
		roots[i].Children = childrenMap[roots[i].ID]
	}

	return roots, nil
}

// CreateLogEntry handles adding a log entry to a project
func (h *Handler) CreateLogEntry(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.GetUsername(r)
	projectID, _ := strconv.Atoi(chi.URLParam(r, "id"))
	note := strings.TrimSpace(r.FormValue("note"))
	url := strings.TrimSpace(r.FormValue("url"))

	if note == "" {
		http.Error(w, "Note is required", http.StatusBadRequest)
		return
	}

	_, err := h.DB.Exec(`
		INSERT INTO log_entries (project_id, username, note, url)
		VALUES (?, ?, ?, ?)`, projectID, username, note, url)

	if err != nil {
		http.Error(w, "Failed to create log entry", http.StatusInternalServerError)
		return
	}

	// Update project timestamp
	h.DB.Exec(`UPDATE projects SET updated_at = CURRENT_TIMESTAMP WHERE id = ?`, projectID)

	// Return updated timeline
	logEntries, _ := h.getLogEntries(projectID, username)
	w.Write([]byte(h.timelineHTML(logEntries, projectID)))
}

// CreateLogReply handles adding a nested reply to a log entry
func (h *Handler) CreateLogReply(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.GetUsername(r)
	projectID, _ := strconv.Atoi(chi.URLParam(r, "id"))
	parentID, _ := strconv.Atoi(chi.URLParam(r, "logId"))
	note := strings.TrimSpace(r.FormValue("note"))
	url := strings.TrimSpace(r.FormValue("url"))

	if note == "" {
		http.Error(w, "Note is required", http.StatusBadRequest)
		return
	}

	_, err := h.DB.Exec(`
		INSERT INTO log_entries (project_id, parent_id, username, note, url)
		VALUES (?, ?, ?, ?, ?)`, projectID, parentID, username, note, url)

	if err != nil {
		http.Error(w, "Failed to create reply", http.StatusInternalServerError)
		return
	}

	h.DB.Exec(`UPDATE projects SET updated_at = CURRENT_TIMESTAMP WHERE id = ?`, projectID)

	logEntries, _ := h.getLogEntries(projectID, username)
	w.Write([]byte(h.timelineHTML(logEntries, projectID)))
}

// DeleteLogEntry handles deleting a log entry
func (h *Handler) DeleteLogEntry(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.GetUsername(r)
	projectID, _ := strconv.Atoi(chi.URLParam(r, "id"))
	logID, _ := strconv.Atoi(chi.URLParam(r, "logId"))

	_, err := h.DB.Exec(`
		DELETE FROM log_entries WHERE id = ? AND username = ?`, logID, username)

	if err != nil {
		http.Error(w, "Failed to delete log entry", http.StatusInternalServerError)
		return
	}

	logEntries, _ := h.getLogEntries(projectID, username)
	w.Write([]byte(h.timelineHTML(logEntries, projectID)))
}

// ReplyForm returns the reply form for a log entry
func (h *Handler) ReplyForm(w http.ResponseWriter, r *http.Request) {
	projectID := chi.URLParam(r, "id")
	logID := chi.URLParam(r, "logId")

	formHTML := fmt.Sprintf(`
		<form hx-post="/projects/%s/log/%s/reply" hx-target="#timeline" hx-swap="innerHTML" class="mt-sm" style="margin-left:24px">
			<input type="text" name="note" placeholder="Add a reply..." required>
			<div class="mt-sm">
				<input type="text" name="url" placeholder="URL (optional)">
			</div>
			<div class="row mt-sm">
				<button type="submit" class="primary">Reply</button>
				<button type="button" onclick="this.closest('form').remove()">Cancel</button>
			</div>
		</form>
	`, projectID, logID)

	w.Write([]byte(formHTML))
}

// --- HTML Rendering ---

func (h *Handler) homeContent(projects []Project, stageFilter, typeFilter string, ratingFilter int) string {
	content := `
		<section class="mb-lg">
			<div class="row space-between mb-md">
				<h2>All Projects</h2>
				<button class="primary" hx-get="/projects/new" hx-target="#modal">New Project</button>
			</div>
	`

	// Filters
	content += `<div class="row mb-md">`

	// Stage filter
	content += `<select name="stage" hx-get="/" hx-target="#projects-table" hx-include="[name='type'],[name='rating']">`
	content += `<option value="">All Stages</option>`
	for _, s := range Stages {
		sel := ""
		if s == stageFilter {
			sel = " selected"
		}
		content += fmt.Sprintf(`<option value="%s"%s>%s</option>`, s, sel, s)
	}
	content += `</select>`

	// Type filter
	content += `<select name="type" hx-get="/" hx-target="#projects-table" hx-include="[name='stage'],[name='rating']">`
	content += `<option value="">All Types</option>`
	types := []struct{ val, label string }{
		{"commercial", "Commercial"},
		{"open-source", "Open Source"},
		{"public", "Public"},
	}
	for _, t := range types {
		sel := ""
		if t.val == typeFilter {
			sel = " selected"
		}
		content += fmt.Sprintf(`<option value="%s"%s>%s</option>`, t.val, sel, t.label)
	}
	content += `</select>`

	// Rating filter
	content += `<select name="rating" hx-get="/" hx-target="#projects-table" hx-include="[name='stage'],[name='type']">`
	content += `<option value="">All Ratings</option>`
	for i := 1; i <= 5; i++ {
		sel := ""
		if i == ratingFilter {
			sel = " selected"
		}
		content += fmt.Sprintf(`<option value="%d"%s>%s</option>`, i, sel, ratingStars(i))
	}
	content += `</select>`

	content += `</div>`

	// Projects table
	content += `<div id="projects-table">`
	content += h.projectsTable(projects)
	content += `</div>`

	content += `</section><div id="modal" class="mt-lg"></div>`
	return content
}

func (h *Handler) projectsTable(projects []Project) string {
	if len(projects) == 0 {
		return `<div class="panel center"><p class="muted">No projects yet. Create your first project!</p></div>`
	}

	content := `<table><thead><tr>
		<th>Name</th><th>Description</th><th>Stage</th><th>Rating</th><th>Flags</th>
	</tr></thead><tbody>`

	for _, p := range projects {
		flags := projectFlags(p)
		content += fmt.Sprintf(`
			<tr onclick="window.location='/projects/%d'" role="link">
				<td>%s</td>
				<td><span class="muted">%s</span></td>
				<td>%s</td>
				<td>%s</td>
				<td><span class="muted">%s</span></td>
			</tr>`,
			p.ID,
			html.EscapeString(p.ShortName),
			html.EscapeString(p.ShortDescription),
			p.Stage,
			ratingDisplay(p.Rating),
			flags)
	}

	content += `</tbody></table>`
	return content
}

func (h *Handler) detailContent(p *Project, logEntries []LogEntry) string {
	content := fmt.Sprintf(`
		<section class="mb-lg">
			<div class="mb-md">
				<a href="/">&larr; Back to projects</a>
			</div>

			<div class="panel mb-md">
				<div class="row space-between mb-sm">
					<h2>%s</h2>
					<div class="row">
						<button hx-get="/projects/%d/edit" hx-target="#modal">Edit</button>
						<button hx-delete="/projects/%d" hx-confirm="Delete this project and all log entries?">Delete</button>
					</div>
				</div>

				<p class="muted mb-sm">%s</p>

				<div class="row mb-sm">
					<span>Stage:</span>
					<select name="stage" hx-put="/projects/%d/stage" hx-swap="none">`,
		html.EscapeString(p.ShortName), p.ID, p.ID,
		html.EscapeString(p.ShortDescription), p.ID)

	for _, s := range Stages {
		sel := ""
		if s == p.Stage {
			sel = " selected"
		}
		content += fmt.Sprintf(`<option value="%s"%s>%s</option>`, s, sel, s)
	}

	content += fmt.Sprintf(`</select>
					<span>Rating: %s</span>
				</div>`, ratingDisplay(p.Rating))

	// URLs
	if p.WebsiteURL != "" {
		content += fmt.Sprintf(`<div class="mb-sm"><span class="muted">Website:</span> <a href="%s" target="_blank">%s</a></div>`,
			html.EscapeString(p.WebsiteURL), html.EscapeString(p.WebsiteURL))
	}
	if p.SourceURL != "" {
		content += fmt.Sprintf(`<div class="mb-sm"><span class="muted">Source:</span> <a href="%s" target="_blank">%s</a></div>`,
			html.EscapeString(p.SourceURL), html.EscapeString(p.SourceURL))
	}

	// Flags
	flags := projectFlags(*p)
	if flags != "" {
		content += fmt.Sprintf(`<div class="mb-sm muted">%s</div>`, flags)
	}

	// Full description
	if p.FullDescription != "" {
		content += fmt.Sprintf(`<div class="mt-md"><p>%s</p></div>`, html.EscapeString(p.FullDescription))
	}

	content += `</div>`

	// Timeline section
	content += fmt.Sprintf(`
		<div class="panel">
			<div class="row space-between mb-md">
				<h2>Timeline</h2>
				<button class="primary" onclick="document.getElementById('new-entry-form').style.display='block'">Add Entry</button>
			</div>

			<form id="new-entry-form" hx-post="/projects/%d/log" hx-target="#timeline" hx-swap="innerHTML" class="mb-md" hidden>
				<input type="text" name="note" placeholder="What happened?" required>
				<div class="mt-sm">
					<input type="text" name="url" placeholder="URL (optional)">
				</div>
				<div class="row mt-sm">
					<button type="submit" class="primary">Add</button>
					<button type="button" onclick="this.closest('form').style.display='none'">Cancel</button>
				</div>
			</form>

			<div id="timeline">
				%s
			</div>
		</div>
	`, p.ID, h.timelineHTML(logEntries, p.ID))

	content += `</section><div id="modal" class="mt-lg"></div>`
	return content
}

func (h *Handler) timelineHTML(entries []LogEntry, projectID int) string {
	if len(entries) == 0 {
		return `<p class="muted center">No entries yet.</p>`
	}

	content := ""
	for _, e := range entries {
		content += h.logEntryHTML(e, projectID, false)
	}
	return content
}

func (h *Handler) logEntryHTML(e LogEntry, projectID int, isChild bool) string {
	indent := ""
	if isChild {
		indent = ` style="margin-left:24px"`
	}

	urlPart := ""
	if e.URL != "" {
		urlPart = fmt.Sprintf(` &mdash; <a href="%s" target="_blank">link</a>`, html.EscapeString(e.URL))
	}

	// Format date - show just the date part
	datePart := e.CreatedAt
	if len(datePart) > 10 {
		datePart = datePart[:10]
	}

	content := fmt.Sprintf(`
		<div class="mb-sm"%s>
			<div class="row space-between">
				<div>
					<span class="muted">%s</span>
					%s%s
				</div>
				<div class="row">
					<button hx-get="/projects/%d/log/%d/reply" hx-target="#reply-%d" hx-swap="innerHTML">Reply</button>
					<button hx-delete="/projects/%d/log/%d" hx-target="#timeline" hx-swap="innerHTML" hx-confirm="Delete this entry?">Del</button>
				</div>
			</div>
			<div id="reply-%d"></div>`,
		indent, datePart,
		html.EscapeString(e.Note), urlPart,
		projectID, e.ID, e.ID,
		projectID, e.ID, e.ID)

	// Render children
	for _, child := range e.Children {
		content += h.logEntryHTML(child, projectID, true)
	}

	content += `</div>`
	return content
}

func (h *Handler) projectForm(p *Project) string {
	title := "New Project"
	action := `hx-post="/projects"`
	shortName := ""
	shortDesc := ""
	fullDesc := ""
	websiteURL := ""
	sourceURL := ""
	isCommercial := ""
	isOpenSource := ""
	isPublic := ""
	stage := "idea"
	rating := 0

	if p != nil {
		title = "Edit Project"
		action = fmt.Sprintf(`hx-put="/projects/%d"`, p.ID)
		shortName = html.EscapeString(p.ShortName)
		shortDesc = html.EscapeString(p.ShortDescription)
		fullDesc = html.EscapeString(p.FullDescription)
		websiteURL = html.EscapeString(p.WebsiteURL)
		sourceURL = html.EscapeString(p.SourceURL)
		if p.IsCommercial {
			isCommercial = " checked"
		}
		if p.IsOpenSource {
			isOpenSource = " checked"
		}
		if p.IsPublic {
			isPublic = " checked"
		}
		stage = p.Stage
		rating = p.Rating
	}

	stageOptions := ""
	for _, s := range Stages {
		sel := ""
		if s == stage {
			sel = " selected"
		}
		stageOptions += fmt.Sprintf(`<option value="%s"%s>%s</option>`, s, sel, s)
	}

	ratingOptions := `<option value="0">No rating</option>`
	for i := 1; i <= 5; i++ {
		sel := ""
		if i == rating {
			sel = " selected"
		}
		ratingOptions += fmt.Sprintf(`<option value="%d"%s>%s</option>`, i, sel, ratingStars(i))
	}

	return fmt.Sprintf(`
		<div class="panel">
			<h2>%s</h2>
			<form %s hx-target="#modal" class="mt-md">
				<label>Name *</label>
				<input type="text" name="short_name" value="%s" required placeholder="Project name">

				<div class="mt-md">
					<label>Tagline</label>
					<input type="text" name="short_description" value="%s" placeholder="Short description (~10 words)">
				</div>

				<div class="mt-md">
					<label>Full Description</label>
					<textarea name="full_description" rows="4" placeholder="Detailed description, goals, notes...">%s</textarea>
				</div>

				<div class="mt-md">
					<label>Website URL</label>
					<input type="url" name="website_url" value="%s" placeholder="https://myproject.com">
				</div>

				<div class="mt-md">
					<label>Source Code URL</label>
					<input type="url" name="source_url" value="%s" placeholder="https://github.com/user/repo">
				</div>

				<div class="mt-md">
					<label>Stage</label>
					<select name="stage">%s</select>
				</div>

				<div class="mt-md">
					<label>Rating</label>
					<select name="rating">%s</select>
				</div>

				<div class="mt-md row">
					<label><input type="checkbox" name="is_commercial"%s> Commercial</label>
					<label><input type="checkbox" name="is_open_source"%s> Open Source</label>
					<label><input type="checkbox" name="is_public"%s> Public</label>
				</div>

				<div class="row mt-md">
					<button type="submit" class="primary">Save</button>
					<button type="button" onclick="this.closest('.panel').remove()">Cancel</button>
				</div>
			</form>
		</div>
	`, title, action, shortName, shortDesc, fullDesc, websiteURL, sourceURL,
		stageOptions, ratingOptions, isCommercial, isOpenSource, isPublic)
}

// --- Helpers ---

func ratingStars(n int) string {
	s := ""
	for i := 0; i < n; i++ {
		s += "★"
	}
	for i := n; i < 5; i++ {
		s += "☆"
	}
	return s
}

func ratingDisplay(n int) string {
	if n == 0 {
		return `<span class="muted">—</span>`
	}
	return ratingStars(n)
}

func projectFlags(p Project) string {
	var flags []string
	if p.IsCommercial {
		flags = append(flags, "commercial")
	}
	if p.IsOpenSource {
		flags = append(flags, "open-source")
	}
	if p.IsPublic {
		flags = append(flags, "public")
	}
	return strings.Join(flags, " · ")
}
