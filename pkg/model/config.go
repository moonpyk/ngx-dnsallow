package model

type Config struct {
	Nginx Nginx   `yaml:"nginx"`
	Hosts []Entry `yaml:"hosts"`
}
