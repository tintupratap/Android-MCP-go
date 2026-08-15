package scrcpy

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type GitHubReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type GitHubRelease struct {
	TagName    string               `json:"tag_name"`
	Prerelease bool                 `json:"prerelease"`
	Draft      bool                 `json:"draft"`
	Assets     []GitHubReleaseAsset `json:"assets"`
}

func FetchLatestRelease(ctx context.Context) (*GitHubRelease, error) {
	url := "https://api.github.com/repos/Genymobile/scrcpy/releases/latest"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for scrcpy releases: %w", err)
	}
	req.Header.Set("User-Agent", "Android-MCP-go")
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch scrcpy release metadata: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned HTTP status %d", resp.StatusCode)
	}

	var rel GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub release response: %w", err)
	}

	return &rel, nil
}

func ResolveAssetForPlatform(rel *GitHubRelease, targetOS, targetArch string) (*GitHubReleaseAsset, error) {
	if targetOS == "" {
		targetOS = runtime.GOOS
	}
	if targetArch == "" {
		targetArch = runtime.GOARCH
	}

	var matchKeywords []string
	switch targetOS {
	case "darwin":
		if targetArch == "arm64" {
			matchKeywords = []string{"macos", "aarch64"}
		} else {
			matchKeywords = []string{"macos", "x86_64"}
		}
	case "windows":
		if targetArch == "386" {
			matchKeywords = []string{"win32"}
		} else {
			matchKeywords = []string{"win64"}
		}
	case "linux":
		matchKeywords = []string{"linux", "x86_64"}
	default:
		return nil, fmt.Errorf("unsupported operating system for scrcpy auto-download: %s", targetOS)
	}

	for _, asset := range rel.Assets {
		name := strings.ToLower(asset.Name)
		matched := true
		for _, kw := range matchKeywords {
			if !strings.Contains(name, kw) {
				matched = false
				break
			}
		}
		if matched {
			return &asset, nil
		}
	}

	return nil, fmt.Errorf("no matching scrcpy release asset found for OS=%s, Arch=%s", targetOS, targetArch)
}

func DownloadFile(ctx context.Context, url, destPath string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download file from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download returned status %d for %s", resp.StatusCode, url)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", destPath, err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write download to %s: %w", destPath, err)
	}

	return nil
}

func ExtractZip(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file %s: %w", zipPath, err)
	}
	defer r.Close()

	destClean := filepath.Clean(destDir)

	for _, f := range r.File {
		filePath := filepath.Join(destDir, f.Name)
		if !strings.HasPrefix(filepath.Clean(filePath), destClean) {
			return fmt.Errorf("illegal zip file path (Zip Slip attack): %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(filePath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("failed to open extracted file %s: %w", filePath, err)
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return fmt.Errorf("failed to open zip entry %s: %w", f.Name, err)
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return fmt.Errorf("failed to extract zip entry %s: %w", f.Name, err)
		}
	}

	return nil
}

func ExtractTarGz(tarPath, destDir string) error {
	f, err := os.Open(tarPath)
	if err != nil {
		return fmt.Errorf("failed to open tar file %s: %w", tarPath, err)
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	destClean := filepath.Clean(destDir)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed reading tar header: %w", err)
		}

		targetPath := filepath.Join(destDir, header.Name)
		if !strings.HasPrefix(filepath.Clean(targetPath), destClean) {
			return fmt.Errorf("illegal tar file path (Tar Slip attack): %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			_ = os.MkdirAll(targetPath, 0755)
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, header.FileInfo().Mode())
			if err != nil {
				return fmt.Errorf("failed creating tar target file %s: %w", targetPath, err)
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return fmt.Errorf("failed copying tar target file %s: %w", targetPath, err)
			}
			outFile.Close()
		}
	}

	return nil
}
