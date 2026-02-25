// cmd/bunkr/selfupdate.go
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/pankajbeniwal/bunkr/internal/ui"
	"github.com/spf13/cobra"
)

const releasesURL = "https://api.github.com/repos/pankajbeniwal/bunkr/releases/latest"

var selfUpdateCmd = &cobra.Command{
	Use:   "self-update",
	Short: "Update bunkr to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		ui.Header("Checking for updates...")

		resp, err := http.Get(releasesURL)
		if err != nil {
			return fmt.Errorf("failed to check for updates: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to check for updates (HTTP %d)", resp.StatusCode)
		}

		var release struct {
			TagName string `json:"tag_name"`
			Assets  []struct {
				Name               string `json:"name"`
				BrowserDownloadURL string `json:"browser_download_url"`
			} `json:"assets"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return fmt.Errorf("failed to parse release info: %w", err)
		}

		if release.TagName == "v"+version {
			ui.Info(fmt.Sprintf("Already at latest version (%s)", version))
			return nil
		}

		ui.Info(fmt.Sprintf("New version available: %s → %s", version, release.TagName))

		// Find matching asset
		arch := runtime.GOARCH
		osName := runtime.GOOS
		assetName := fmt.Sprintf("bunkr_%s_%s", osName, arch)

		var downloadURL string
		for _, asset := range release.Assets {
			if asset.Name == assetName {
				downloadURL = asset.BrowserDownloadURL
				break
			}
		}

		if downloadURL == "" {
			return fmt.Errorf("no binary found for %s/%s", osName, arch)
		}

		// Download
		ui.Info("Downloading...")
		dlResp, err := http.Get(downloadURL)
		if err != nil {
			return fmt.Errorf("failed to download: %w", err)
		}
		defer dlResp.Body.Close()

		// Get current binary path
		execPath, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to determine executable path: %w", err)
		}

		// Write to temp file first
		tmpFile, err := os.CreateTemp("", "bunkr-update-*")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := io.Copy(tmpFile, dlResp.Body); err != nil {
			tmpFile.Close()
			return fmt.Errorf("failed to write update: %w", err)
		}
		tmpFile.Close()

		if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
			return fmt.Errorf("failed to set permissions: %w", err)
		}

		// Replace binary
		if err := os.Rename(tmpFile.Name(), execPath); err != nil {
			return fmt.Errorf("failed to replace binary: %w", err)
		}

		ui.Result(fmt.Sprintf("Updated to %s", release.TagName))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(selfUpdateCmd)
}
