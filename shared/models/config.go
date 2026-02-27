package models

// AppConfig represents the application configuration
type AppConfig struct {
	AppName        string           `yaml:"app_name"`
	AppVersion     string           `yaml:"app_version"`
	AppReleaseDate string           `yaml:"app_release_date"`
	ChangelogPath  string           `yaml:"changelog_path"`
	Port           int              `yaml:"port"`
	DatabaseURL    string           `yaml:"database_url"`
	JWTSecret      string           `yaml:"jwt_secret"`
	Users          []UserFromConfig `yaml:"users"`
}
