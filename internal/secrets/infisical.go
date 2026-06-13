// Purpose: Load secrets from Infisical CLI or environment variables.
// Docs: internal/secrets/infisical.doc.md
package secrets

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// InfisicalClient loads secrets from Infisical CLI.
type InfisicalClient struct {
	projectID string
	env       string
}

// NewInfisicalClient creates a client with optional defaults.
func NewInfisicalClient() *InfisicalClient {
	return &InfisicalClient{
		projectID: getEnvDefault("INFISICAL_PROJECT_ID", "fa7758b4-f84c-4297-966e-710056d531ef"),
		env:       getEnvDefault("INFISICAL_ENV", "dev"),
	}
}

// LoadSerpAPIKeys loads SerpAPI keys from Infisical or env.
func (c *InfisicalClient) LoadSerpAPIKeys() []string {
	keys := c.loadFromInfisical([]string{"SERPAPI_KEY_1", "SERPAPI_KEY_2", "SERPAPI_KEY_3", "SERPAPI_KEY_4"})
	if len(keys) > 0 {
		return keys
	}
	var envKeys []string
	for i := 1; i <= 4; i++ {
		if k := os.Getenv(fmt.Sprintf("SERPAPI_KEY_%d", i)); k != "" {
			envKeys = append(envKeys, k)
		}
	}
	return envKeys
}

// LoadAllSecrets loads all known secrets and sets them in the process env.
func (c *InfisicalClient) LoadAllSecrets() map[string]string {
	names := []string{
		"SERPAPI_KEY_1", "SERPAPI_KEY_2", "SERPAPI_KEY_3", "SERPAPI_KEY_4",
		"BRAVE_API_KEY", "OPENROUTER_API_KEY", "SCRAPECREATORS_API_KEY",
		"GROQ_API_KEY", "OPENAI_API_KEY", "SIN_WEBSEARCH_TOKEN",
	}
	result := make(map[string]string)
	for _, name := range names {
		if val := c.getSecret(name); val != "" {
			result[name] = val
			os.Setenv(name, val)
		}
	}
	return result
}

func (c *InfisicalClient) loadFromInfisical(names []string) []string {
	var keys []string
	for _, name := range names {
		if val := c.getSecret(name); val != "" {
			keys = append(keys, val)
		}
	}
	return keys
}

func (c *InfisicalClient) getSecret(name string) string {
	if _, err := exec.LookPath("infisical"); err != nil {
		return ""
	}
	if val := os.Getenv(name); val != "" {
		return val
	}
	cmd := exec.Command("infisical", "secrets", "get", name,
		"--projectId="+c.projectID, "--env="+c.env, "--plain", "--silent")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// IsAvailable checks if the Infisical CLI is installed and authenticated.
func (c *InfisicalClient) IsAvailable() bool {
	if _, err := exec.LookPath("infisical"); err != nil {
		return false
	}
	cmd := exec.Command("infisical", "whoami", "--silent")
	return cmd.Run() == nil
}

// Status returns diagnostic info.
func (c *InfisicalClient) Status() map[string]interface{} {
	path, _ := exec.LookPath("infisical")
	return map[string]interface{}{
		"available":  c.IsAvailable(),
		"project_id": c.projectID,
		"env":        c.env,
		"cli_path":   path,
	}
}

// ExportAsJSON dumps all loaded secrets as redacted JSON.
func (c *InfisicalClient) ExportAsJSON() (string, error) {
	secrets := c.LoadAllSecrets()
	redacted := make(map[string]string)
	for k, v := range secrets {
		if len(v) > 8 {
			redacted[k] = v[:4] + strings.Repeat("*", len(v)-8) + v[len(v)-4:]
		} else {
			redacted[k] = "***"
		}
	}
	data, err := json.MarshalIndent(redacted, "", "  ")
	return string(data), err
}

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
