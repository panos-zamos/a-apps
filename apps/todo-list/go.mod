module github.com/panos-zamos/a-apps/apps/todo-list

go 1.21

require (
	github.com/go-chi/chi/v5 v5.0.11
	github.com/panos-zamos/a-apps/shared/auth v0.0.0
	github.com/panos-zamos/a-apps/shared/database v0.0.0
	github.com/panos-zamos/a-apps/shared/models v0.0.0
	github.com/panos-zamos/a-apps/shared/templates v0.0.0
)

require (
	github.com/golang-jwt/jwt/v5 v5.2.0 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	golang.org/x/crypto v0.19.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/panos-zamos/a-apps/shared/auth => ../../shared/auth

replace github.com/panos-zamos/a-apps/shared/database => ../../shared/database

replace github.com/panos-zamos/a-apps/shared/models => ../../shared/models

replace github.com/panos-zamos/a-apps/shared/templates => ../../shared/templates
