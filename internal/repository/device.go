// Package repository provides implementations of the domain repository interfaces.
package repository

import (
	"os"

	"github.com/kainnsoft/duoscans/internal/connection"
	"github.com/kainnsoft/duoscans/internal/domain"
	"github.com/kainnsoft/duoscans/internal/downloader"
)

// DeviceRepository implements domain.ScreenshotRepository by pulling files
// from a connected device via a Connector.
type DeviceRepository struct {
	dl *downloader.Downloader
}

// NewDeviceRepository creates a DeviceRepository using conn to transfer files
// and tempDir as the local staging area.
func NewDeviceRepository(conn connection.Connector, tempDir string) *DeviceRepository {
	return &DeviceRepository{dl: downloader.New(conn, tempDir)}
}

// Fetch downloads up to numFiles screenshots from galleryPath on the device.
func (r *DeviceRepository) Fetch(galleryPath string, numFiles int) ([]domain.Screenshot, error) {
	results, err := r.dl.Download(galleryPath, numFiles)
	if err != nil {
		return nil, err
	}

	screenshots := make([]domain.Screenshot, len(results))
	for i, res := range results {
		screenshots[i] = domain.Screenshot{
			LocalPath:  res.LocalPath,
			RemotePath: res.RemotePath,
		}
	}
	return screenshots, nil
}

// Cleanup removes the local temp files created by Fetch.
func (r *DeviceRepository) Cleanup(screenshots []domain.Screenshot) {
	for _, s := range screenshots {
		os.Remove(s.LocalPath)
	}
}
