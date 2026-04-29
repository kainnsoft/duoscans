// Package usecase contains application-level orchestration logic.
package usecase

import (
	"fmt"

	"github.com/kainnsoft/duoscans/internal/domain"
)

// DownloadScreenshotsUseCase downloads screenshots from a connected device.
type DownloadScreenshotsUseCase struct {
	Screenshots domain.ScreenshotRepository
}

// DownloadScreenshotsInput holds the parameters for a single run.
type DownloadScreenshotsInput struct {
	GalleryPath string
	NumFiles    int
}

// Execute downloads screenshots and returns the number of downloaded files.
func (uc *DownloadScreenshotsUseCase) Execute(input DownloadScreenshotsInput) (int, error) {
	results, err := uc.Screenshots.Fetch(input.GalleryPath, input.NumFiles)
	if err != nil {
		return 0, fmt.Errorf("fetch screenshots: %w", err)
	}
	return len(results), nil
}
