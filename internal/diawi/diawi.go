package diawi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type uploadResponse struct {
	Job string `json:"job"`
}

type statusResponse struct {
	Status int    `json:"status"`
	Link   string `json:"link"`
	Hash   string `json:"hash"`
}

// Upload uploads a file to Diawi and returns the install link.
func Upload(token, filePath string) (string, error) {
	// 1. Upload file
	job, err := uploadFile(token, filePath)
	if err != nil {
		return "", fmt.Errorf("diawi upload: %w", err)
	}
	fmt.Println("  Uploaded, processing...")

	// 2. Poll status
	for i := 0; i < 60; i++ {
		time.Sleep(3 * time.Second)
		link, done, err := checkStatus(token, job)
		if err != nil {
			return "", err
		}
		if done {
			return link, nil
		}
		fmt.Print(".")
	}
	fmt.Println()

	return "", fmt.Errorf("diawi processing timeout")
}

func uploadFile(token, filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	_ = w.WriteField("token", token)

	fw, err := w.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(fw, file); err != nil {
		return "", err
	}
	w.Close()

	resp, err := http.Post("https://upload.diawi.com/", w.FormDataContentType(), &buf)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result uploadResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("diawi response: %s", string(body))
	}
	if result.Job == "" {
		return "", fmt.Errorf("diawi error: %s", string(body))
	}
	return result.Job, nil
}

func checkStatus(token, job string) (string, bool, error) {
	url := fmt.Sprintf("https://upload.diawi.com/status?token=%s&job=%s", token, job)
	resp, err := http.Get(url)
	if err != nil {
		return "", false, err
	}
	defer resp.Body.Close()

	var result statusResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", false, err
	}

	switch result.Status {
	case 2000: // OK
		return result.Link, true, nil
	case 2001: // processing
		return "", false, nil
	default:
		return "", false, fmt.Errorf("diawi error, status: %d", result.Status)
	}
}
