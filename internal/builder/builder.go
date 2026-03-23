package builder

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Result struct {
	FilePath string
	FileName string
}

// flavorToScheme maps Android flavor to iOS scheme.
var flavorToScheme = map[string]string{
	"develop":    "Development",
	"production": "Production",
}

// BuildAPK runs flutter build apk for the given flavor.
func BuildAPK(flavor string) (*Result, error) {
	cmd := exec.Command("fvm", "flutter", "build", "apk", "--flavor", flavor)
	cmd.Dir = ".."
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("flutter build apk failed: %w", err)
	}

	apkDir := filepath.Join("..", "build", "app", "outputs", "flutter-apk")
	entries, err := os.ReadDir(apkDir)
	if err != nil {
		return nil, fmt.Errorf("read apk dir: %w", err)
	}

	suffix := fmt.Sprintf("%s-release.apk", flavor)
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), suffix) {
			fullPath, _ := filepath.Abs(filepath.Join(apkDir, e.Name()))
			return &Result{
				FilePath: fullPath,
				FileName: e.Name(),
			}, nil
		}
	}

	return nil, fmt.Errorf("APK not found for flavor %s in %s", flavor, apkDir)
}

// BuildIPA runs flutter build ipa for the given flavor.
func BuildIPA(flavor string) (*Result, error) {
	scheme := flavorToScheme[flavor]
	if scheme == "" {
		scheme = flavor
	}

	cmd := exec.Command("fvm", "flutter", "build", "ipa", "--flavor", scheme)
	cmd.Dir = ".."
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("flutter build ipa failed: %w", err)
	}

	ipaDir := filepath.Join("..", "build", "ios", "ipa")
	entries, err := os.ReadDir(ipaDir)
	if err != nil {
		return nil, fmt.Errorf("read ipa dir: %w", err)
	}

	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".ipa") {
			fullPath, _ := filepath.Abs(filepath.Join(ipaDir, e.Name()))
			return &Result{
				FilePath: fullPath,
				FileName: e.Name(),
			}, nil
		}
	}

	return nil, fmt.Errorf("IPA not found in %s", ipaDir)
}
