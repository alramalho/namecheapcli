package namecheap

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ConfigFromEnv(sandbox bool) Config {
	config := configFromFile(DefaultConfigPath())
	applyEnv(&config)
	if config.Endpoint == "" && sandbox {
		config.Endpoint = SandboxEndpoint
	}
	return config
}

func DefaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, ".namecheapcli")
}

func WriteConfigFile(path string, config Config) error {
	if path == "" {
		return fmt.Errorf("config path is empty")
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err := fmt.Fprintf(file, "NAMECHEAP_API_USER=%s\n", config.APIUser); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(file, "NAMECHEAP_API_KEY=%s\n", config.APIKey); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(file, "NAMECHEAP_USERNAME=%s\n", config.UserName); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(file, "NAMECHEAP_CLIENT_IP=%s\n", config.ClientIP); err != nil {
		return err
	}
	if config.Endpoint != "" {
		if _, err := fmt.Fprintf(file, "NAMECHEAP_ENDPOINT=%s\n", config.Endpoint); err != nil {
			return err
		}
	}
	return nil
}

func configFromFile(path string) Config {
	if path == "" {
		return Config{}
	}
	file, err := os.Open(path)
	if err != nil {
		return Config{}
	}
	defer file.Close()

	values := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		if key != "" {
			values[key] = value
		}
	}

	return Config{
		Endpoint: first(values, "NAMECHEAP_ENDPOINT", "ENDPOINT"),
		APIUser:  first(values, "NAMECHEAP_API_USER", "API_USER"),
		APIKey:   first(values, "NAMECHEAP_API_KEY", "API_KEY"),
		UserName: first(values, "NAMECHEAP_USERNAME", "USERNAME"),
		ClientIP: first(values, "NAMECHEAP_CLIENT_IP", "CLIENT_IP"),
	}
}

func applyEnv(config *Config) {
	if value := strings.TrimSpace(os.Getenv("NAMECHEAP_ENDPOINT")); value != "" {
		config.Endpoint = value
	}
	if value := strings.TrimSpace(os.Getenv("NAMECHEAP_API_USER")); value != "" {
		config.APIUser = value
	}
	if value := strings.TrimSpace(os.Getenv("NAMECHEAP_API_KEY")); value != "" {
		config.APIKey = value
	}
	if value := strings.TrimSpace(os.Getenv("NAMECHEAP_USERNAME")); value != "" {
		config.UserName = value
	}
	if value := strings.TrimSpace(os.Getenv("NAMECHEAP_CLIENT_IP")); value != "" {
		config.ClientIP = value
	}
}

func first(values map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(values[key]); value != "" {
			return value
		}
	}
	return ""
}
