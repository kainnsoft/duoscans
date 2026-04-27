// Package builder is the composition root: it wires together infrastructure,
// repositories, and use-cases, then runs the application.
package builder

import (
	"fmt"

	"github.com/kainnsoft/duoscans/config"
	"github.com/kainnsoft/duoscans/internal/connection"
	"github.com/kainnsoft/duoscans/internal/ocr"
	"github.com/kainnsoft/duoscans/internal/repository"
	"github.com/kainnsoft/duoscans/internal/usecase"
	commonconfig "github.com/kainnsoft/duoscans/pkg/config"
)

const (
	commandDownload       = "download"
	commandFindDuplicates = "find-duplicates"
)

// Run loads configuration and executes one CLI command.
func Run(args []string) error {
	if len(args) > 0 && args[0] == "--" {
		args = args[1:]
	}

	if len(args) == 0 {
		return fmt.Errorf("missing command; usage: duoscans [%s|%s]", commandDownload, commandFindDuplicates)
	}

	cfg, err := commonconfig.FromEnv()
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	switch args[0] {
	case commandDownload:
		return runDownload(&cfg)
	case commandFindDuplicates:
		return runFindDuplicates(&cfg)
	default:
		return fmt.Errorf("unknown command %q; usage: duoscans [%s|%s]", args[0], commandDownload, commandFindDuplicates)
	}
}

func runDownload(cfg *config.Config) error {
	conn, err := newConnector(cfg)
	if err != nil {
		return fmt.Errorf("create connector: %w", err)
	}
	if err := conn.Connect(); err != nil {
		return fmt.Errorf("connect to device: %w", err)
	}
	defer conn.Close()

	uc := &usecase.DownloadScreenshotsUseCase{
		Screenshots: repository.NewDeviceRepository(conn, cfg.TempDir),
	}
	n, err := uc.Execute(usecase.DownloadScreenshotsInput{
		GalleryPath: cfg.GalleryPath,
		NumFiles:    cfg.NumFiles,
	})
	if err != nil {
		return fmt.Errorf("execute: %w", err)
	}

	fmt.Printf("downloaded %d files\n", n)
	if n > 0 {
		fmt.Printf("files are saved under %s\n", cfg.TempDir)
	}

	return nil
}

func runFindDuplicates(cfg *config.Config) error {
	ocrSvc := ocr.New()
	defer ocrSvc.Close()

	logRepo, err := repository.NewCSVLogRepository(cfg.LogFile)
	if err != nil {
		return fmt.Errorf("init log repository: %w", err)
	}

	uc := &usecase.FindDuplicatesUseCase{
		Screenshots: repository.NewLocalDirectoryRepository(cfg.TempDir),
		OCR:         ocrSvc,
		Log:         logRepo,
	}
	n, err := uc.Execute(usecase.FindDuplicatesInput{
		GalleryPath: cfg.GalleryPath,
		NumFiles:    cfg.NumFiles,
	})
	if err != nil {
		return fmt.Errorf("execute: %w", err)
	}

	if n > 0 {
		fmt.Printf("results written to %s\n", cfg.LogFile)
	}

	return nil
}

// newConnector selects the Connector implementation based on cfg.ConnectionType.
func newConnector(cfg *config.Config) (connection.Connector, error) {
	switch cfg.ConnectionType {
	case config.ConnectionADB:
		return &connection.ADB{Serial: cfg.DeviceAddress}, nil
	case config.ConnectionUSB:
		return &connection.USB{DevicePath: cfg.DeviceAddress}, nil
	case config.ConnectionWiFi:
		return &connection.WiFi{Address: cfg.DeviceAddress}, nil
	default:
		return nil, fmt.Errorf("unknown connection type: %s", cfg.ConnectionType)
	}
}
