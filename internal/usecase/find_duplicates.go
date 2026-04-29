// Package usecase contains application-level orchestration logic.
package usecase

import (
	"fmt"
	"os"

	"github.com/kainnsoft/duoscans/internal/domain"
)

// FindDuplicatesUseCase orchestrates the duplicate-detection pipeline:
// read local screenshots → extract text → find duplicates → log results.
type FindDuplicatesUseCase struct {
	Screenshots domain.ScreenshotRepository
	OCR         domain.OCRService
	Log         domain.DuplicateLogRepository
}

// FindDuplicatesInput holds the parameters for a single run.
type FindDuplicatesInput struct {
	GalleryPath string
	NumFiles    int
}

// Execute runs the pipeline and returns the number of duplicate groups found.
func (uc *FindDuplicatesUseCase) Execute(input FindDuplicatesInput) (int, error) {
	raw, err := uc.Screenshots.Fetch(input.GalleryPath, input.NumFiles)
	if err != nil {
		return 0, fmt.Errorf("fetch screenshots: %w", err)
	}
	//defer uc.Screenshots.Cleanup(raw)

	fmt.Printf("loaded %d files from temp_dir\n", len(raw))

	var enriched []domain.Screenshot
	for _, s := range raw {
		text, err := uc.OCR.ExtractText(s.LocalPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: OCR failed for %s: %v\n", s.LocalPath, err)
			continue
		}
		s.Sentence = text
		enriched = append(enriched, s)
	}

	groups := domain.FindDuplicates(enriched)
	fmt.Printf("found %d duplicate groups\n", len(groups))

	if len(groups) == 0 {
		return 0, nil
	}

	if err := uc.Log.Write(groups); err != nil {
		return 0, fmt.Errorf("write log: %w", err)
	}

	return len(groups), nil
}
