package namecheap

import (
	"os"
	"strings"
)

func ConfigFromEnv(sandbox bool) Config {
	endpoint := strings.TrimSpace(os.Getenv("NAMECHEAP_ENDPOINT"))
	if endpoint == "" && sandbox {
		endpoint = SandboxEndpoint
	}
	return Config{
		Endpoint: endpoint,
		APIUser:  strings.TrimSpace(os.Getenv("NAMECHEAP_API_USER")),
		APIKey:   strings.TrimSpace(os.Getenv("NAMECHEAP_API_KEY")),
		UserName: strings.TrimSpace(os.Getenv("NAMECHEAP_USERNAME")),
		ClientIP: strings.TrimSpace(os.Getenv("NAMECHEAP_CLIENT_IP")),
	}
}
