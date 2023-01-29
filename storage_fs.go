package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// FSStorage is a file system storage implementation.
type FSStorage struct {
	root string
}

const (
	storageFileName = "data.json"
	imagesDirName   = "images"
)

// NewFSStorage returns a new File Storage instance.
func NewFSStorage(root string) Storage {
	return &FSStorage{
		root: root,
	}
}

// Put stores an apartment in the file system storage.
func (s *FSStorage) Put(apt Apartment) error {
	if err := ensureDir(s.root, 0775); err != nil {
		return fmt.Errorf("ensure root dir: %w", err)
	}

	address := strings.ReplaceAll(apt.Address, " ", "_")
	address = strings.ToLower(address)

	dir := fmt.Sprintf("%s/%s_%d", s.root, address, apt.ID)
	if err := ensureDir(dir, 0755); err != nil {
		return fmt.Errorf("ensure apartment dir: %w", err)
	}

	// Write apartment data.
	f, err := os.Create(fmt.Sprintf("%s/%s", dir, storageFileName))
	if err != nil {
		return fmt.Errorf("create storage file: %w", err)
	}
	defer f.Close()

	data, err := json.Marshal(apt)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	if _, err := f.Write(data); err != nil {
		return fmt.Errorf("write: %w", err)
	}

	d := strings.Join([]string{dir, imagesDirName}, "/")
	if err := ensureDir(d, 0775); err != nil {
		return fmt.Errorf("ensure images dir: %w", err)
	}

	for i, url := range apt.ImageURLs {
		downloaded, err := downloadImage(url, d)
		if err != nil {
			return fmt.Errorf("download image: %w", err)
		}

		if !downloaded {
			fmt.Printf("Skipped %s (%d/%d)\n", url, i+1, len(apt.ImageURLs))
			continue
		}
		fmt.Printf("Downloaded %s (%d/%d)\n", url, i+1, len(apt.ImageURLs))
	}
	return nil
}

func downloadImage(url, dir string) (bool, error) {
	fileName := fmt.Sprintf("%s/%s", dir, url[strings.LastIndex(url, "/")+1:])

	// Ensure folder path exist
	if err := os.MkdirAll(filepath.Dir(fileName), os.ModePerm); err != nil {
		return false, fmt.Errorf("mkdirall: %w", err)
	}

	// Skip if the file is already downloaded.
	// os.Stat runs normally if file exists.
	// it's expected behaviour for os.Stat, so it doesn't throw an error
	_, err := os.Stat(fileName)
	if err == nil {
		return false, nil
	}

	f, err := os.Create(fileName)
	if err != nil {
		return false, fmt.Errorf("create: %w", err)
	}
	defer f.Close()

	resp, err := http.Get(url)
	if err != nil {
		return false, fmt.Errorf("get: %w", err)
	}
	defer resp.Body.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return false, fmt.Errorf("copy: %w", err)
	}

	return true, nil
}
