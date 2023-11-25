package internal

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// DownloadImage downloads an image from the URL and save it to the local file system.
func DownloadImage(url string, id string, number int) {
	// Expect like this: ./static/images/
	imagesPath := ImagesPath
	base := filepath.Join(imagesPath, id)
	dest := filepath.Join(base, fmt.Sprintf("%d", number)+".png")

	// Create directory
	if _, err := os.Stat(base); os.IsNotExist(err) {
		Logger.Info("Creating directory: " + base)
		err := os.MkdirAll(base, 0777)
		if err != nil {
			panic(err)
		}
	}

	// Prepare a new file
	Logger.Info("Downloading image: " + url)
	file, err := os.Create(dest)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Download image
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Authorization", "token "+GitHubToken)
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Check response
	contentType := resp.Header.Get("Content-Type")
	if resp.StatusCode != 200 || contentType != "image/png" {
		Logger.Error(fmt.Sprintf("Response: %d %s", resp.StatusCode, contentType))

		// Remove the file
		err := os.Remove(dest)
		if err != nil {
			panic(err)
		}

		return
	}
	Logger.Info(fmt.Sprintf("Response: %d %s", resp.StatusCode, contentType))

	// Write the body to file
	written, err := io.Copy(file, resp.Body)
	if err != nil {
		panic(err)
	}
	Logger.Info("Downloaded image: " + dest + " (" + fmt.Sprintf("%d", written) + " bytes)")
}
