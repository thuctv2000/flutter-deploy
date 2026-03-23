package version

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Bump reads and bumps the version in pubspec.yaml.
func Bump() (string, error) {
	content, err := os.ReadFile("../pubspec.yaml")
	if err != nil {
		return "", fmt.Errorf("read pubspec.yaml: %w", err)
	}
	lines := strings.Split(string(content), "\n")
	var newVersion string
	for i, line := range lines {
		if after, ok := strings.CutPrefix(line, "version:"); ok {
			current := strings.TrimSpace(after)
			parts := strings.Split(current, "+")
			versionName := parts[0]
			oldBuild := parts[1]
			toDay := time.Now().Format("060102")
			oldDate := oldBuild[:6]
			oldSeq := oldBuild[6:]
			var newSeq int
			if oldDate == toDay {
				seq, _ := strconv.Atoi(oldSeq)
				newSeq = seq + 1
			} else {
				newSeq = 1
			}
			newBuild := fmt.Sprintf("%s%02d", toDay, newSeq)
			newVersion = fmt.Sprintf("%s+%s", versionName, newBuild)
			lines[i] = "version: " + newVersion
			fmt.Println("New version:", newVersion)
			break
		}
	}
	if newVersion == "" {
		return "", fmt.Errorf("version not found in pubspec.yaml")
	}
	output := strings.Join(lines, "\n")
	return newVersion, os.WriteFile("../pubspec.yaml", []byte(output), 0644)
}
