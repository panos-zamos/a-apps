# From Zero to Deployed: Building a Personal App Platform in One Afternoon

*February 17, 2026*

This is a step-by-step guide to building your own multi-app platform using Go, HTMX, and SQLite. By the end, you'll have:

- A working shopping list app
- A scaffold script to generate new apps in minutes
- Deployment ready for a $6/month server
- Authentication built-in
- A structure optimized for LLM-assisted development

Total time: ~4 hours if you're following along. Let's go.

## Part 1: The Vision (5 minutes)

You want to build multiple small web apps—a grocery list, expense tracker, project notes, whatever. You could:

1. **Separate repos** - Too much overhead, repeated auth code
2. **Microservices** - Overengineered for personal use
3. **All-in-one app** - Feature soup, hard to isolate

Instead: **Monorepo with independent apps sharing utilities**

Each app:
- Has its own SQLite database
- Runs on its own port
- Shares auth and templates
- Can be modified without breaking others

## Part 2: Foundation (30 minutes)

### Create the Structure

```bash
mkdir a-apps && cd a-apps
mkdir -p apps shared/{auth,database,templates,models} deploy scripts docs
```

### Go Workspace

Create `go.work`:
```go
go 1.21

use (
    ./shared/auth
    ./shared/database
    ./shared/templates
    ./shared/models
)
```

This lets Go modules reference each other during development.

### Shared Models

`shared/models/go.mod`:
```go
module github.com/yourusername/a-apps/shared/models
go 1.21
```

`shared/models/user.go`:
```go
package models

type UserFromConfig struct {
    Username     string `yaml:"username"`
    PasswordHash string `yaml:"password_hash"`
}
```

### Database Package

`shared/database/go.mod`:
```go
module github.com/yourusername/a-apps/shared/database
go 1.21

require github.com/mattn/go-sqlite3 v1.14.22
```

`shared/database/sqlite.go`:
```go
package database

import (
    "database/sql"
    "fmt"
    _ "github.com/mattn/go-sqlite3"
)

type DB struct {
    *sql.DB
    Path string
}

func Open(path string) (*DB, error) {
    db, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
    if err != nil {
        return nil, err
    }
    return &DB{DB: db, Path: path}, nil
}

func (db *DB) RunMigrations(migrations []string) error {
    _, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS migrations (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT NOT NULL UNIQUE,
            applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
        )
    `)
    if err != nil {
        return err
    }

    for i, migration := range migrations {
        name := fmt.Sprintf("migration_%03d", i+1)
        
        var count int
        db.QueryRow("SELECT COUNT(*) FROM migrations WHERE name = ?", name).Scan(&count)
        
        if count > 0 {
            continue
        }

        if _, err := db.Exec(migration); err != nil {
            return fmt.Errorf("migration %s failed: %w", name, err)
        }

        db.Exec("INSERT INTO migrations (name) VALUES (?)", name)
    }

    return nil
}
```

### Auth Package

`shared/auth/go.mod`:
```go
module github.com/yourusername/a-apps/shared/auth
go 1.21

require (
    github.com/golang-jwt/jwt/v5 v5.2.0
    golang.org/x/crypto v0.19.0
    gopkg.in/yaml.v3 v3.0.1
)
```

`shared/auth/auth.go`:
```go
package auth

import (
    "fmt"
    "os"
    "time"
    
    "github.com/golang-jwt/jwt/v5"
    "golang.org/x/crypto/bcrypt"
    "gopkg.in/yaml.v3"
)

type UserFromConfig struct {
    Username     string `yaml:"username"`
    PasswordHash string `yaml:"password_hash"`
}

func LoadUsersFromConfig(path string) ([]UserFromConfig, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var config struct {
        Users []UserFromConfig `yaml:"users"`
    }

    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, err
    }

    return config.Users, nil
}

func ValidateCredentials(username, password string, users []UserFromConfig) (bool, error) {
    for _, user := range users {
        if user.Username == username {
            err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
            return err == nil, err
        }
    }
    return false, fmt.Errorf("user not found")
}

func GenerateToken(username, secret string) (string, error) {
    claims := jwt.MapClaims{
        "username": username,
        "exp":      time.Now().Add(24 * time.Hour).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

func ValidateToken(tokenString, secret string) (string, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return []byte(secret), nil
    })

    if err != nil {
        return "", err
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        username := claims["username"].(string)
        return username, nil
    }

    return "", fmt.Errorf("invalid token")
}
```

`shared/auth/middleware.go`:
```go
package auth

import (
    "context"
    "net/http"
)

type contextKey string

const UserContextKey contextKey = "username"

func Middleware(jwtSecret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            cookie, err := r.Cookie("auth_token")
            if err != nil {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
            }

            username, err := ValidateToken(cookie.Value, jwtSecret)
            if err != nil {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
            }

            ctx := context.WithValue(r.Context(), UserContextKey, username)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func GetUsername(r *http.Request) (string, bool) {
    username, ok := r.Context().Value(UserContextKey).(string)
    return username, ok
}
```

### Templates Package

`shared/templates/base.go`:
```go
package templates

const BaseHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body class="bg-gray-50 min-h-screen">
    <nav class="bg-white shadow-sm">
        <div class="max-w-7xl mx-auto px-4 py-4 flex justify-between">
            <h1 class="text-xl font-bold">{{.AppName}}</h1>
            {{if .Username}}
            <div>
                <span class="mr-4">{{.Username}}</span>
                <form action="/logout" method="POST" style="display:inline">
                    <button type="submit">Logout</button>
                </form>
            </div>
            {{end}}
        </div>
    </nav>
    <main class="max-w-7xl mx-auto py-6 px-4">
        {{.Content}}
    </main>
</body>
</html>`

const LoginHTML = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Login</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-50 flex items-center justify-center min-h-screen">
    <div class="bg-white p-8 rounded-lg shadow max-w-md">
        <h2 class="text-2xl font-bold mb-4">{{.AppName}}</h2>
        {{if .Error}}
        <div class="bg-red-50 text-red-700 p-3 rounded mb-4">{{.Error}}</div>
        {{end}}
        <form method="POST" action="/login">
            <input type="text" name="username" placeholder="Username" required
                   class="w-full px-3 py-2 border rounded mb-3" />
            <input type="password" name="password" placeholder="Password" required
                   class="w-full px-3 py-2 border rounded mb-3" />
            <button type="submit" class="w-full bg-indigo-600 text-white py-2 rounded">
                Sign In
            </button>
        </form>
    </div>
</body>
</html>`
```

### Install Dependencies

```bash
cd shared/database && go mod tidy
cd ../auth && go mod tidy
cd ../..
```

## Part 3: Your First App (45 minutes)

### Create App Structure

```bash
mkdir -p apps/shopping-list/{handlers,data}
cd apps/shopping-list
```

### go.mod

```go
module github.com/yourusername/a-apps/apps/shopping-list
go 1.21

require (
    github.com/go-chi/chi/v5 v5.0.11
    github.com/yourusername/a-apps/shared/auth v0.0.0
    github.com/yourusername/a-apps/shared/database v0.0.0
    github.com/yourusername/a-apps/shared/templates v0.0.0
)

replace github.com/yourusername/a-apps/shared/auth => ../../shared/auth
replace github.com/yourusername/a-apps/shared/database => ../../shared/database
replace github.com/yourusername/a-apps/shared/templates => ../../shared/templates
```

### Database Migrations

`handlers/migrations.go`:
```go
package handlers

var Migrations = []string{
    `CREATE TABLE IF NOT EXISTS stores (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        username TEXT NOT NULL,
        color TEXT DEFAULT '#3B82F6',
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
}
```

### Main Handlers

`handlers/handler.go`:
```go
package handlers

import (
    "fmt"
    "html/template"
    "net/http"
    
    "github.com/yourusername/a-apps/shared/auth"
    "github.com/yourusername/a-apps/shared/database"
    sharedTemplates "github.com/yourusername/a-apps/shared/templates"
)

type Handler struct {
    DB        *database.DB
    Users     []auth.UserFromConfig
    JWTSecret string
}

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
    tmpl := template.Must(template.New("login").Parse(sharedTemplates.LoginHTML))
    tmpl.Execute(w, map[string]interface{}{
        "AppName": "Shopping List",
        "Error":   r.URL.Query().Get("error"),
    })
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    username := r.FormValue("username")
    password := r.FormValue("password")

    valid, _ := auth.ValidateCredentials(username, password, h.Users)
    if !valid {
        http.Redirect(w, r, "/login?error=Invalid credentials", http.StatusSeeOther)
        return
    }

    token, _ := auth.GenerateToken(username, h.JWTSecret)

    http.SetCookie(w, &http.Cookie{
        Name:     "auth_token",
        Value:    token,
        Path:     "/",
        HttpOnly: true,
        MaxAge:   86400,
    })

    http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
    http.SetCookie(w, &http.Cookie{
        Name:   "auth_token",
        Value:  "",
        Path:   "/",
        MaxAge: -1,
    })
    http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
    username, _ := auth.GetUsername(r)
    
    data := map[string]interface{}{
        "Title":    "Shopping List",
        "AppName":  "Shopping List",
        "Username": username,
        "Content":  template.HTML("<h2>Welcome! Start adding stores.</h2>"),
    }

    tmpl := template.Must(template.New("base").Parse(sharedTemplates.BaseHTML))
    tmpl.Execute(w, data)
}
```

### Main Application

`main.go`:
```go
package main

import (
    "log"
    "net/http"
    "os"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    "github.com/yourusername/a-apps/apps/shopping-list/handlers"
    "github.com/yourusername/a-apps/shared/auth"
    "github.com/yourusername/a-apps/shared/database"
)

func main() {
    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        jwtSecret = "dev-secret-change-in-production"
    }

    port := os.Getenv("PORT")
    if port == "" {
        port = "3001"
    }

    db, err := database.Open("data/shopping-list.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    if err := db.RunMigrations(handlers.Migrations); err != nil {
        log.Fatal(err)
    }

    users, _ := auth.LoadUsersFromConfig("config.yaml")

    h := &handlers.Handler{
        DB:        db,
        Users:     users,
        JWTSecret: jwtSecret,
    }

    r := chi.NewRouter()
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

    r.Get("/login", h.LoginPage)
    r.Post("/login", h.Login)
    r.Post("/logout", h.Logout)

    r.Group(func(r chi.Router) {
        r.Use(auth.Middleware(jwtSecret))
        r.Get("/", h.Home)
    })

    log.Printf("Starting on :%s", port)
    http.ListenAndServe(":"+port, r)
}
```

### Configuration

`config.yaml`:
```yaml
users:
  - username: demo
    password_hash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
    # password: demo123
```

### Update Workspace

Back in root directory:
```bash
cd ../..
echo "./apps/shopping-list" >> go.work
```

### Run It!

```bash
cd apps/shopping-list
go mod tidy
go build
./shopping-list
```

Visit http://localhost:3001
- Username: `demo`
- Password: `demo123`

🎉 **You have a working authenticated app!**

## Part 4: Scaffold Script (30 minutes)

Now automate app creation.

`scripts/new-app.sh`:
```bash
#!/bin/bash
set -e

if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <app-name> <port>"
    exit 1
fi

APP_NAME=$1
APP_PORT=$2

echo "Creating $APP_NAME on port $APP_PORT..."

mkdir -p "apps/$APP_NAME/handlers"

# Copy shopping-list as template
cp -r apps/shopping-list/. "apps/$APP_NAME/"

# Replace names
find "apps/$APP_NAME" -type f -exec sed -i "s/shopping-list/$APP_NAME/g" {} \;
sed -i "s/3001/$APP_PORT/g" "apps/$APP_NAME/main.go"
sed -i "s/Shopping List/${APP_NAME^}/g" "apps/$APP_NAME/handlers/handler.go"

# Add to workspace
echo "./apps/$APP_NAME" >> go.work

echo "✓ Created! Run with: cd apps/$APP_NAME && go run main.go"
```

```bash
chmod +x scripts/new-app.sh
```

### Test It

```bash
./scripts/new-app.sh expense-tracker 3002
cd apps/expense-tracker
go mod tidy
go run main.go
```

Visit http://localhost:3002 - another app running!

## Part 5: Deployment (1 hour)

### Docker Setup

`apps/shopping-list/Dockerfile`:
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/config.yaml .
RUN mkdir -p data
EXPOSE 3001
CMD ["./main"]
```

### Docker Compose

`deploy/docker-compose.yml`:
```yaml
version: '3.8'

services:
  shopping-list:
    build:
      context: ../apps/shopping-list
    container_name: shopping-list
    environment:
      - PORT=3001
      - JWT_SECRET=${JWT_SECRET}
    volumes:
      - ../apps/shopping-list/data:/root/data
    restart: unless-stopped
    networks:
      - apps

  nginx:
    image: nginx:alpine
    container_name: nginx
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - shopping-list
    restart: unless-stopped
    networks:
      - apps

networks:
  apps:
```

### Nginx Config

`deploy/nginx.conf`:
```nginx
events {
    worker_connections 1024;
}

http {
    server {
        listen 80;

        location /shopping {
            rewrite ^/shopping/(.*)$ /$1 break;
            proxy_pass http://shopping-list:3001;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }

        location / {
            return 200 "Apps Platform\n";
        }
    }
}
```

### Test Locally

```bash
cd deploy
JWT_SECRET=mysecret docker-compose up --build
```

Visit http://localhost/shopping

### Deploy to Digital Ocean

1. **Create Droplet** ($6/month basic)
2. **Install Docker**
3. **rsync your code**

```bash
./scripts/deploy.sh
```

`scripts/deploy.sh`:
```bash
#!/bin/bash
SERVER=root@your-server-ip

rsync -avz --exclude='data' --exclude='.git' ./ $SERVER:/opt/a-apps/

ssh $SERVER << 'EOF'
cd /opt/a-apps/deploy
export JWT_SECRET="$(openssl rand -base64 32)"
docker-compose down
docker-compose up --build -d
docker-compose ps
EOF

echo "✓ Deployed!"
```

## Part 6: Next Steps (ongoing)

### Add More Apps

```bash
./scripts/new-app.sh notes 3003
./scripts/new-app.sh bookmarks 3004
```

### Enhance with LLM

Ask Claude/GPT:
> "In apps/notes, add rich text editing with Tiptap. Reference shopping-list for the handler structure."

### Monitoring

Add a simple dashboard:
```bash
./scripts/new-app.sh dashboard 3000
# Shows all apps, health checks, DB sizes
```

### Backups

`scripts/backup.sh`:
```bash
#!/bin/bash
tar -czf backups/$(date +%Y%m%d).tar.gz apps/*/data/*.db
```

Cron job:
```bash
0 2 * * * /opt/a-apps/scripts/backup.sh
```

## What You've Built

In ~4 hours:

✅ Multi-app platform with shared utilities  
✅ SQLite databases (cheap, simple backups)  
✅ JWT authentication  
✅ HTMX for dynamic UIs  
✅ Docker deployment  
✅ Scaffold script for new apps  
✅ Nginx reverse proxy  
✅ Production-ready setup on $6/month

**Total cost:** $6/month for unlimited apps.

## Common Issues

**"Module not found"** - Run `go mod tidy` in the app directory

**"Can't connect to DB"** - Check `data/` directory exists

**"Invalid credentials"** - Generate new hash: `htpasswd -bnBC 10 "" password`

**"Port already in use"** - Kill other process or use different port

## Resources

- Full code: [github.com/yourusername/a-apps](https://github.com)
- HTMX docs: [htmx.org](https://htmx.org)
- Chi router: [github.com/go-chi/chi](https://github.com/go-chi/chi)

---

You now have a platform that can grow with you. Add expense tracking, habit monitoring, recipe management—whatever you need. Each app is isolated but benefits from shared auth and deployment.

Most startups don't need more than this. Why should your side projects?

*Questions? Find me [@yourhandle](https://twitter.com) or open an issue on GitHub!*
