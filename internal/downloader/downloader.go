// Package downloader pulls screenshot files from a device to a local temp directory.
package downloader

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kainnsoft/duoscans/internal/connection"
)

const duolingoFileMarker = "Duo"

// Result holds a successfully downloaded file's local path and its remote origin.
type Result struct {
	LocalPath  string
	RemotePath string
}

// Downloader fetches screenshots from a device via a Connector.
type Downloader struct {
	conn    connection.Connector
	tempDir string
}

// New creates a Downloader that saves files to tempDir.
func New(conn connection.Connector, tempDir string) *Downloader {
	return &Downloader{conn: conn, tempDir: tempDir}
}

// Download pulls up to numFiles screenshots from remoteDir.
// Returns a slice of Results for every file successfully downloaded.
func (d *Downloader) Download(remoteDir string, numFiles int) ([]Result, error) {
	if err := os.MkdirAll(d.tempDir, 0o755); err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}

	paths, err := d.conn.ListFiles(remoteDir, 0)
	if err != nil {
		return nil, fmt.Errorf("list remote files: %w", err)
	}

	var filtered []string
	for _, remote := range paths {
		if strings.Contains(filepath.Base(remote), duolingoFileMarker) {
			filtered = append(filtered, remote)
		}
	}

	// Screenshot names include a time component, so descending filename order
	// gives newest files first for supported naming patterns.
	sort.Slice(filtered, func(i, j int) bool {
		return filepath.Base(filtered[i]) > filepath.Base(filtered[j])
	})

	if numFiles > 0 && len(filtered) > numFiles {
		filtered = filtered[:numFiles]
	}

	var results []Result
	for _, remote := range filtered {
		local := filepath.Join(d.tempDir, filepath.Base(remote))
		if err := d.conn.Pull(remote, local); err != nil {
			// Log and continue — a single bad file should not abort the run.
			fmt.Fprintf(os.Stderr, "warning: skipping %s: %v\n", remote, err)
			continue
		}
		results = append(results, Result{LocalPath: local, RemotePath: remote})
	}
	return results, nil
}

// Cleanup removes all files that were downloaded to the temp directory.
func (d *Downloader) Cleanup(results []Result) {
	for _, r := range results {
		os.Remove(r.LocalPath)
	}
}
