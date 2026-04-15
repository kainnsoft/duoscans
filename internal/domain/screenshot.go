// Package domain contains core entities, value objects, and the interfaces
// that use-cases depend on (repository and service contracts).
package domain

// Screenshot represents a single Duolingo screenshot pulled from the device.
type Screenshot struct {
	LocalPath  string // temporary local file path after download
	RemotePath string // original path on the device
	Sentence   string // English sentence extracted by OCR
}

// DuplicateGroup holds a sentence and all screenshots that share it.
type DuplicateGroup struct {
	Sentence    string
	Screenshots []Screenshot
}

// ScreenshotRepository fetches screenshots from the connected device.
type ScreenshotRepository interface {
	// Fetch downloads up to numFiles screenshots from galleryPath and returns
	// them as Screenshot values with LocalPath and RemotePath populated.
	// Callers are responsible for cleanup of local files.
	Fetch(galleryPath string, numFiles int) ([]Screenshot, error)

	// Cleanup removes the local temp files created by Fetch.
	Cleanup(screenshots []Screenshot)
}

// OCRService extracts text from an image file.
type OCRService interface {
	ExtractText(imagePath string) (string, error)
	Close()
}

// DuplicateLogRepository persists duplicate groups to a durable store.
type DuplicateLogRepository interface {
	Write(groups []DuplicateGroup) error
}

// FindDuplicates is pure domain logic: it groups screenshots by sentence and
// returns only groups that contain more than one screenshot.
func FindDuplicates(screenshots []Screenshot) []DuplicateGroup {
	index := make(map[string][]Screenshot)
	for _, s := range screenshots {
		if s.Sentence == "" {
			continue
		}
		index[s.Sentence] = append(index[s.Sentence], s)
	}

	var groups []DuplicateGroup
	for sentence, members := range index {
		if len(members) > 1 {
			groups = append(groups, DuplicateGroup{Sentence: sentence, Screenshots: members})
		}
	}
	return groups
}
