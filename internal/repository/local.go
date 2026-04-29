package repository

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kainnsoft/duoscans/internal/domain"
)

const duolingoFileMarker = "Duo"

// LocalDirectoryRepository reads screenshots from a local directory.
type LocalDirectoryRepository struct {
	dir string
}

// NewLocalDirectoryRepository creates a repository over a local directory.
func NewLocalDirectoryRepository(dir string) *LocalDirectoryRepository {
	return &LocalDirectoryRepository{dir: dir}
}

// Fetch reads up to numFiles image files from dir and returns local screenshots.
func (r *LocalDirectoryRepository) Fetch(_ string, numFiles int) ([]domain.Screenshot, error) {
	entries, err := os.ReadDir(r.dir)
	if err != nil {
		return nil, fmt.Errorf("read temp dir %q: %w", r.dir, err)
	}

	var paths []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name))
		switch ext {
		case ".jpg", ".jpeg", ".png", ".webp", ".bmp", ".tif", ".tiff", ".heic":
			if strings.Contains(name, duolingoFileMarker) {
				paths = append(paths, filepath.Join(r.dir, name))
			}
		}
	}

	// Keep order deterministic and process most recent naming patterns first.
	sort.Slice(paths, func(i, j int) bool {
		return filepath.Base(paths[i]) > filepath.Base(paths[j])
	})

	if numFiles > 0 && len(paths) > numFiles {
		paths = paths[:numFiles]
	}

	out := make([]domain.Screenshot, 0, len(paths))
	for _, p := range paths {
		out = append(out, domain.Screenshot{LocalPath: p})
	}
	return out, nil
}

// Cleanup is a no-op for local files; this repository does not own them.
func (r *LocalDirectoryRepository) Cleanup(_ []domain.Screenshot) {}
