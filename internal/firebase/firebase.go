package firebase

import (
	"bytes"
	"fmt"
	"flutter-deploy/internal/config"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type UploadResult struct {
	ConsoleLink  string
	DownloadLink string
}

// Upload distributes the APK via Firebase App Distribution using service account.
func Upload(cfg *config.FirebaseEnv, apkPath, releaseNotes string) (*UploadResult, error) {
	args := []string{
		"appdistribution:distribute", apkPath,
		"--app", cfg.AppID,
	}
	if releaseNotes != "" {
		args = append(args, "--release-notes", releaseNotes)
	}
	if cfg.Groups != "" {
		args = append(args, "--groups", cfg.Groups)
	}

	saPath, _ := filepath.Abs(cfg.ServiceAccountPath)

	cmd := exec.Command("firebase", args...)
	cmd.Dir = ".."
	cmd.Env = append(os.Environ(), "GOOGLE_APPLICATION_CREDENTIALS="+saPath)

	var buf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &buf)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("firebase upload failed: %w", err)
	}

	result := &UploadResult{}
	for _, line := range strings.Split(buf.String(), "\n") {
		if strings.Contains(line, "console.firebase.google.com") {
			result.ConsoleLink = extractURL(line)
		}
		if strings.Contains(line, "appdistribution.firebase.google.com") {
			result.ConsoleLink = extractURL(line)
		}
		if strings.Contains(line, "firebaseappdistribution.googleapis.com") {
			result.DownloadLink = extractURL(line)
		}
	}

	return result, nil
}

func extractURL(line string) string {
	start := strings.Index(line, "https://")
	if start == -1 {
		return ""
	}
	url := line[start:]
	// trim trailing whitespace or control chars
	url = strings.TrimSpace(url)
	return url
}
