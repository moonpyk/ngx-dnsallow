package model

// Config represents the main configuration structure for the application.
type Config struct {
	// Nginx holds the configuration related to nginx.
	Nginx Nginx `yaml:"nginx"`
	// Hosts is a list of host entries.
	Hosts []Entry `yaml:"hosts"`
}
