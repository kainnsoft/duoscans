# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

`duoscans` connects to an Android smartphone, downloads Duolingo screenshots from the Gallery, extracts English sentences via OCR (Tesseract), finds duplicates, and writes results to a CSV log. No files are ever deleted from the device.

## System prerequisites

```bash
sudo apt install tesseract-ocr          # required by gosseract
adb                                     # required for ADB/Wi-Fi connections
```

## Commands

```bash
# Build
go build ./...

# Run (uses config.json by default)
go run . -config config.json

# Test
go test ./...

# Run a single test
go test ./internal/duplicate/... -run TestFind

# Lint (requires golangci-lint)
golangci-lint run
```

## Architecture

```
main.go                        # wires everything together: load config → connect → download → OCR → find duplicates → log
config/config.go               # JSON config (connection_type, num_files, gallery_path, temp_dir, log_file)
internal/connection/           # Connector interface + ADB (implemented), USB and Wi-Fi (stubs)
internal/downloader/           # pulls remote files to temp dir, cleans up after run
internal/ocr/                  # gosseract wrapper; normalises extracted text to lowercase single-line
internal/duplicate/            # groups []File by sentence, returns groups with >1 member
internal/logger/               # appends duplicate groups to a CSV (creates header on first run)
```

## Data flow

1. `config.Load` reads `config.json` (falls back to defaults if missing)
2. A `connection.Connector` is chosen based on `connection_type` (`adb` | `usb` | `wifi`)
3. `downloader.Download` calls `ListFiles` (up to `num_files`) then `Pull` for each → local temp files
4. `ocr.Extractor.ExtractText` runs Tesseract on each file; text is lowercased and whitespace-collapsed
5. `duplicate.Find` groups files by sentence; only groups with ≥2 files are returned
6. `logger.CSVLogger.Write` appends rows to the CSV (`timestamp, sentence, file_count, local_path, remote_path`)
7. Temp files are removed via `downloader.Cleanup`

## Config file reference

| Field | Default | Description |
|---|---|---|
| `connection_type` | `"adb"` | `"adb"`, `"usb"`, or `"wifi"` |
| `device_address` | `""` | Serial (ADB/USB) or `IP:port` (Wi-Fi) |
| `gallery_path` | `/sdcard/DCIM/Screenshots` | Remote path on device |
| `num_files` | `50` | How many files to download per run |
| `temp_dir` | system temp | Local directory for downloaded files |
| `log_file` | `duplicates.csv` | Output CSV path |

## Key implementation notes

- USB and Wi-Fi connectors are stubs — only ADB is functional.
- `ocr.normalise` lowercases and collapses whitespace. If two screenshots differ only in punctuation, they will still be treated as different; adjust `normalise` if that becomes an issue.
- The CSV logger appends; it does not overwrite. Each run adds new rows with an RFC3339 timestamp.
