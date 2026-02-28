package config

import "os"

type Config struct {
       Token          string
       Username       string
       RefreshSeconds int
}

func Load() *Config {
       sec := 60
       if v := os.Getenv("GITHUB_REFRESH_SECONDS"); v != "" {
	       if n, err := strconv.Atoi(v); err == nil && n > 0 {
		       sec = n
	       }
       }
       return &Config{
	       Token:          os.Getenv("GITHUB_TOKEN"),
	       Username:       os.Getenv("GITHUB_USERNAME"),
	       RefreshSeconds: sec,
       }
}
