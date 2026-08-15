package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/tintupratap/Android-MCP-go/internal/config"
	"github.com/tintupratap/Android-MCP-go/internal/logging"
)

const SkillsRepoBaseURL = "https://raw.githubusercontent.com/tintupratap/Android-MCP-go/main/skills"

type SkillDomain struct {
	ID          string `json:"id"`
	File        string `json:"file"`
	Description string `json:"description"`
}

type Manifest struct {
	Name        string        `json:"name"`
	Version     string        `json:"version"`
	Description string        `json:"description"`
	Domains     []SkillDomain `json:"domains"`
}

type SkillItem struct {
	Name        string `json:"name"`
	Command     string `json:"command,omitempty"`
	MCPTool     string `json:"mcp_tool,omitempty"`
	Status      string `json:"status"`
	Description string `json:"description,omitempty"`
}

type DomainManifest struct {
	Domain      string      `json:"domain"`
	Description string      `json:"description"`
	Skills      []SkillItem `json:"skills"`
}

type Manager struct {
	paths *config.RuntimePaths
}

func NewManager() (*Manager, error) {
	paths, err := config.GetRuntimePaths()
	if err != nil {
		return nil, err
	}
	return &Manager{paths: paths}, nil
}

func (m *Manager) SkillsDir() string {
	return filepath.Join(m.paths.Root, "skills")
}

func (m *Manager) ManifestPath() string {
	return filepath.Join(m.SkillsDir(), "manifest.json")
}

func (m *Manager) IsInstalled() bool {
	if _, err := os.Stat(m.ManifestPath()); err == nil {
		return true
	}
	return false
}

var EmbeddedManifest = `{
  "name": "Android-MCP-go",
  "version": "0.4.0",
  "description": "Production-grade Go implementation of the Model Context Protocol server for Android operating systems",
  "domains": [
    { "id": "device", "file": "device.json", "description": "Device management, discovery, and wireless bootstrap" },
    { "id": "ui", "file": "ui.json", "description": "UI hierarchy parsing, selectors, and element interaction" },
    { "id": "screenshot", "file": "screenshot.json", "description": "Screen capture and visual annotations" },
    { "id": "input", "file": "input.json", "description": "Touch gestures, text typing, and key events" },
    { "id": "applications", "file": "applications.json", "description": "Package listing, application launching, stopping, and package info" },
    { "id": "filesystem", "file": "filesystem.json", "description": "Android file push, pull, directory listing, and removal" },
    { "id": "shell", "file": "shell.json", "description": "ADB shell execution with context timeouts" },
    { "id": "automation", "file": "automation.json", "description": "Workflow execution and verification assertions" },
    { "id": "scrcpy", "file": "scrcpy.json", "description": "Managed scrcpy display mirroring and process management" },
    { "id": "platform_tools", "file": "platform_tools.json", "description": "Self-contained Android SDK Platform-Tools management" }
  ]
}`

func (m *Manager) EnsureSkills(ctx context.Context) error {
	skillsDir := m.SkillsDir()
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		return err
	}

	manifestPath := m.ManifestPath()
	if _, err := os.Stat(manifestPath); err != nil {
		if err := os.WriteFile(manifestPath, []byte(EmbeddedManifest), 0644); err != nil {
			return fmt.Errorf("failed to write embedded manifest.json: %w", err)
		}
	}

	// Copy/Download domain manifests
	var manifest Manifest
	data, err := os.ReadFile(manifestPath)
	if err == nil {
		_ = json.Unmarshal(data, &manifest)
	}

	for _, domain := range manifest.Domains {
		targetFile := filepath.Join(skillsDir, filepath.Base(domain.File))
		if _, err := os.Stat(targetFile); err != nil {
			// Download from repository
			url := fmt.Sprintf("%s/%s", SkillsRepoBaseURL, filepath.Base(domain.File))
			if err := downloadSkillFile(ctx, url, targetFile); err != nil {
				logging.Debugf("Skill file download skipped for %s: %v", domain.ID, err)
			}
		}
	}

	return nil
}

func (m *Manager) ListSkills() string {
	var sb strings.Builder
	sb.WriteString("Android-MCP-go Installed Skills\n")
	sb.WriteString("===============================\n\n")

	skillsDir := m.SkillsDir()
	sb.WriteString(fmt.Sprintf("Skills Directory: %s\n\n", skillsDir))

	manifestPath := m.ManifestPath()
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		sb.WriteString(fmt.Sprintf("Status: Not installed (%v)\n", err))
		return sb.String()
	}

	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		sb.WriteString(fmt.Sprintf("Status: Corrupt manifest (%v)\n", err))
		return sb.String()
	}

	sb.WriteString(fmt.Sprintf("Version: %s\n", manifest.Version))
	sb.WriteString(fmt.Sprintf("Total Capability Domains: %d\n\n", len(manifest.Domains)))

	for _, domain := range manifest.Domains {
		sb.WriteString(fmt.Sprintf("Domain [%s]: %s\n", domain.ID, domain.Description))

		domainPath := filepath.Join(skillsDir, filepath.Base(domain.File))
		if domainData, err := os.ReadFile(domainPath); err == nil {
			var dm DomainManifest
			if err := json.Unmarshal(domainData, &dm); err == nil {
				for _, sk := range dm.Skills {
					name := sk.Name
					if sk.MCPTool != "" {
						name = fmt.Sprintf("%s (%s)", sk.Name, sk.MCPTool)
					}
					sb.WriteString(fmt.Sprintf("  - %-35s [%s]\n", name, sk.Status))
				}
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func downloadSkillFile(ctx context.Context, url, dest string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
