package repository

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/kainnsoft/duoscans/internal/domain"
)

// CSVLogRepository implements domain.DuplicateLogRepository by appending rows
// to a CSV file. The file is created with a header if it does not yet exist.
type CSVLogRepository struct {
	path string
}

// NewCSVLogRepository returns a CSVLogRepository that writes to path.
func NewCSVLogRepository(path string) (*CSVLogRepository, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		f, err := os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("create log file: %w", err)
		}
		defer f.Close()
		if _, err := fmt.Fprintln(f, "timestamp,sentence,file_count,local_path,remote_path"); err != nil {
			return nil, err
		}
	}
	return &CSVLogRepository{path: path}, nil
}

// Write appends one row per screenshot in each duplicate group.
func (r *CSVLogRepository) Write(groups []domain.DuplicateGroup) error {
	f, err := os.OpenFile(r.path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	ts := time.Now().UTC().Format(time.RFC3339)

	for _, g := range groups {
		for _, s := range g.Screenshots {
			if err := w.Write([]string{
				ts,
				g.Sentence,
				fmt.Sprintf("%d", len(g.Screenshots)),
				s.LocalPath,
				s.RemotePath,
			}); err != nil {
				return err
			}
		}
	}
	w.Flush()
	return w.Error()
}
