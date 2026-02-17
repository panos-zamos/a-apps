module github.com/panos/a-apps/shared/auth

go 1.21

require (
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/panos/a-apps/shared/models v0.0.0
	golang.org/x/crypto v0.19.0
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/panos/a-apps/shared/models => ../models
