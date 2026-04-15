// Package ocr extracts text from images using Tesseract via gosseract.
package ocr

import (
	"strings"

	"github.com/otiai10/gosseract/v2"
)

// Extractor reads text from image files.
type Extractor struct {
	client *gosseract.Client
}

// New creates an Extractor. Call Close when done.
func New() *Extractor {
	client := gosseract.NewClient()
	client.SetLanguage("eng")
	return &Extractor{client: client}
}

// Close releases the Tesseract client.
func (e *Extractor) Close() {
	e.client.Close()
}

// ExtractText returns the text found in the image at imagePath.
// The result is trimmed and normalised to a single line per sentence.
func (e *Extractor) ExtractText(imagePath string) (string, error) {
	if err := e.client.SetImage(imagePath); err != nil {
		return "", err
	}
	text, err := e.client.Text()
	if err != nil {
		return "", err
	}
	return normalise(text), nil
}

// normalise collapses whitespace and trims the string so two screenshots
// with identical sentences compare equal regardless of minor OCR noise.
func normalise(s string) string {
	lines := strings.Fields(s) // split on any whitespace, drop empty tokens
	return strings.ToLower(strings.Join(lines, " "))
}
