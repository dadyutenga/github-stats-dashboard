package config

import "os"

type Config struct {
	Token    string
	Username string
}

func Load() *Config {
	return &Config{
		Token:    os.Getenv("GITHUB_TOKEN"),
		Username: os.Getenv("GITHUB_USERNAME"), // optional override
	}
}
