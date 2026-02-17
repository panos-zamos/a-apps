package models

// AppConfig represents the application configuration
type AppConfig struct {
	AppName     string           `yaml:"app_name"`
	Port        int              `yaml:"port"`
	DatabaseURL string           `yaml:"database_url"`
	JWTSecret   string           `yaml:"jwt_secret"`
	Users       []UserFromConfig `yaml:"users"`
}
