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

	// Check if this is an HTMX request for just the card list
	if r.Header.Get("HX-Request") == "true" {
		w.Write([]byte(h.projectCards(projects)))
		return
	}

	data := map[string]interface{}{
		"Title":          "Projects",
		"AppName":        "Projects",
		"Username":       username,
		"Content":        template.HTML(h.homeContent(projects, stageFilter, typeFilter, ratingFilter)),
		"AppVersion":     h.AppConfig.AppVersion,
		"AppReleaseDate": h.AppConfig.AppReleaseDate,
		"ChangelogURL":   "/changelog",
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
		"Title":          project.ShortName,
		"AppName":        "Projects",
		"Username":       username,
		"Content":        template.HTML(h.detailContent(project, logEntries)),
		"AppVersion":     h.AppConfig.AppVersion,
		"AppReleaseDate": h.AppConfig.AppReleaseDate,
		"ChangelogURL":   "/changelog",
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
		<form hx-post="/projects/%s/log/%s/reply" hx-target="#timeline" hx-swap="innerHTML" class="mt-sm">
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
		<div class="row space-between mb-md">
			<button class="btn" hx-get="/projects/new" hx-target="#modal">add project</button>
		</div>
	`

	// Filters
	content += `<div class="filter-bar mb-md">`

	// All filter
	activeClass := ""
	if stageFilter == "" && typeFilter == "" && ratingFilter == 0 {
		activeClass = " active"
	}
	content += fmt.Sprintf(`<button class="chip%s" hx-get="/?stage=&type=&rating=" hx-target="#projects-list">all</button>`, activeClass)

	// Stage filter
	for _, s := range Stages {
		check := ""
		if s == stageFilter {
			check = " active"
		}
		content += fmt.Sprintf(`<button class="chip%s" hx-get="/?stage=%s&type=%s&rating=%d" hx-target="#projects-list">%s</button>`, check, s, typeFilter, ratingFilter, s)
	}

	content += `</div>`

	// Project list
	content += `<div id="projects-list" class="list">`
	content += h.projectCards(projects)
	content += `</div>`

	content += `<div id="modal" class="mt-lg"></div>`
	return content
}

func (h *Handler) projectCards(projects []Project) string {
	if len(projects) == 0 {
		return `<div class="panel center"><p class="muted">No projects yet. Create your first project!</p></div>`
	}

	content := ""
	for _, p := range projects {
		flags := projectFlags(p)
		content += fmt.Sprintf(`
			<a href="/projects/%d" class="list-item">
				<div class="list-item-top">
					<span class="item-name">%s</span>
					<span class="stage %s">%s</span>
				</div>
				<p class="item-desc">%s</p>
				<div class="list-item-bottom">
					<span class="rating">%s</span>
					<span class="flags">%s</span>
				</div>
			</a>`,
			p.ID,
			html.EscapeString(p.ShortName),
			badgeClass(p.Stage), p.Stage,
			html.EscapeString(p.ShortDescription),
			ratingDisplay(p.Rating),
			flags)
	}

	return content
}

func (h *Handler) detailContent(p *Project, logEntries []LogEntry) string {
	content := `<a href="/" class="back">← back</a>`

	// Title + badge
	content += fmt.Sprintf(`
		<div class="list-item-top mb-sm">
			<h2>%s</h2>
			<span class="stage %s">%s</span>
		</div>`,
		html.EscapeString(p.ShortName), badgeClass(p.Stage), p.Stage)

	// Tagline
	if p.ShortDescription != "" {
		content += fmt.Sprintf(`<p class="muted mb-md">%s</p>`, html.EscapeString(p.ShortDescription))
	}

	// Metadata
	content += `<table class="meta-table">`

	// Stage
	content += fmt.Sprintf(`<tr><td>stage</td><td><span class="stage %s">%s</span></td></tr>`, badgeClass(p.Stage), p.Stage)

	// Rating
	content += fmt.Sprintf(`<tr><td>rating</td><td><span class="rating">%s</span></td></tr>`, ratingDisplay(p.Rating))

	// Website
	if p.WebsiteURL != "" {
		content += fmt.Sprintf(`<tr><td>website</td><td><a href="%s" target="_blank">%s</a></td></tr>`,
			html.EscapeString(p.WebsiteURL), html.EscapeString(p.WebsiteURL))
	}

	// Source
	if p.SourceURL != "" {
		content += fmt.Sprintf(`<tr><td>source</td><td><a href="%s" target="_blank">%s</a></td></tr>`,
			html.EscapeString(p.SourceURL), html.EscapeString(p.SourceURL))
	}

	// Flags
	flags := projectFlags(*p)
	if flags != "" {
		content += fmt.Sprintf(`<tr><td>flags</td><td>%s</td></tr>`, flags)
	}

	content += `</table>`

	// Action buttons
	content += fmt.Sprintf(`
		<div class="row mb-md">
			<button class="btn" hx-get="/projects/%d/edit" hx-target="#modal">Edit</button>
			<button class="btn btn-danger" hx-delete="/projects/%d" hx-confirm="Delete this project and all log entries?">Delete</button>
		</div>`, p.ID, p.ID)

	// Full description
	if p.FullDescription != "" {
		content += fmt.Sprintf(`
			<div class="section-label mb-md">description</div>
			<p class="mb-lg">%s</p>`, html.EscapeString(p.FullDescription))
	}

	// Timeline section
	content += fmt.Sprintf(`
		<div class="section-label mb-md">timeline</div>
		<div class="mb-md">
			<button class="btn" onclick="document.getElementById('new-entry-form').style.display='block'">Add Entry</button>
		</div>

		<form id="new-entry-form" hx-post="/projects/%d/log" hx-target="#timeline" hx-swap="innerHTML" class="panel mb-md" hidden>
			<div class="field">
				<label>note</label>
				<input type="text" name="note" placeholder="What happened?" required>
			</div>
			<div class="field">
				<label>url</label>
				<input type="text" name="url" placeholder="https://...">
			</div>
			<div class="row mt-md">
				<button type="submit" class="btn btn-pop">Save</button>
				<button type="button" class="btn" onclick="this.closest('form').style.display='none'">Cancel</button>
			</div>
		</form>

		<div id="timeline" class="timeline">
			%s
		</div>
	`, p.ID, h.timelineHTML(logEntries, p.ID))

	content += `<div id="modal" class="mt-lg"></div>`
	return content
}

func (h *Handler) timelineHTML(entries []LogEntry, projectID int) string {
	if len(entries) == 0 {
		return `<p class="muted center">No entries yet.</p>`
	}

	content := ""
	for _, e := range entries {
		content += h.logEntryHTML(e, projectID, false)
		for _, child := range e.Children {
			content += h.logEntryHTML(child, projectID, true)
		}
	}
	return content
}

func (h *Handler) logEntryHTML(e LogEntry, projectID int, isChild bool) string {
	nestedClass := ""
	if isChild {
		nestedClass = " nested"
	}

	urlPart := ""
	if e.URL != "" {
		urlPart = fmt.Sprintf(` — <a href="%s" target="_blank">link</a>`, html.EscapeString(e.URL))
	}

	datePart := e.CreatedAt
	if len(datePart) > 10 {
		datePart = datePart[:10]
	}

	content := fmt.Sprintf(`
		<article class="entry%s">
			<div class="entry-date">%s</div>
			<div class="entry-text">%s%s</div>
			<div class="entry-actions">
				<button hx-get="/projects/%d/log/%d/reply" hx-target="#reply-%d" hx-swap="innerHTML">reply</button>
				<button hx-delete="/projects/%d/log/%d" hx-target="#timeline" hx-swap="innerHTML" hx-confirm="Delete this entry?">delete</button>
			</div>
			<div id="reply-%d"></div>
		</article>`,
		nestedClass, datePart,
		html.EscapeString(e.Note), urlPart,
		projectID, e.ID, e.ID,
		projectID, e.ID, e.ID)

	return content
}

func (h *Handler) projectForm(p *Project) string {
	title := "new project"
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
		title = "edit project"
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

	ratingOptions := `<option value="0">no rating</option>`
	for i := 1; i <= 5; i++ {
		sel := ""
		if i == rating {
			sel = " selected"
		}
		ratingOptions += fmt.Sprintf(`<option value="%d"%s>%d/5</option>`, i, sel, i)
	}

	return fmt.Sprintf(`
		<div class="panel">
			<h3>%s</h3>
			<form %s hx-target="#modal" class="mt-md">
				<div class="field">
					<label>name *</label>
					<input type="text" name="short_name" value="%s" required placeholder="project name">
				</div>

				<div class="field">
					<label>tagline</label>
					<input type="text" name="short_description" value="%s" placeholder="short description">
				</div>

				<div class="field">
					<label>description</label>
					<textarea name="full_description" placeholder="detailed description...">%s</textarea>
				</div>

				<div class="field">
					<label>website url</label>
					<input type="url" name="website_url" value="%s" placeholder="https://...">
				</div>

				<div class="field">
					<label>source code url</label>
					<input type="url" name="source_url" value="%s" placeholder="https://github.com/user/repo">
				</div>

				<div class="field">
					<label>stage</label>
					<select name="stage">%s</select>
				</div>

				<div class="field">
					<label>rating</label>
					<select name="rating">%s</select>
				</div>

				<div class="field">
					<div class="checkbox-row">
						<label><input type="checkbox" name="is_commercial"%s> commercial</label>
						<label><input type="checkbox" name="is_open_source"%s> open source</label>
						<label><input type="checkbox" name="is_public"%s> public</label>
					</div>
				</div>

				<div class="row mt-md">
					<button type="submit" class="btn btn-pop">Save</button>
					<button type="button" class="btn" onclick="this.closest('.panel').remove()">Cancel</button>
				</div>
			</form>
		</div>
	`, title, action, shortName, shortDesc, fullDesc, websiteURL, sourceURL,
		stageOptions, ratingOptions, isCommercial, isOpenSource, isPublic)
}

// --- Helpers ---

func badgeClass(stage string) string {
	switch stage {
	case "development":
		return "stage-dev"
	default:
		return "stage-" + stage
	}
}

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
		return `<span class="rating empty-rating">—</span>`
	}
	html := `<span class="rating">`
	for i := 0; i < 5; i++ {
		if i < n {
			html += `<span class="pip on"></span>`
		} else {
			html += `<span class="pip"></span>`
		}
	}
	html += `</span>`
	return html
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
