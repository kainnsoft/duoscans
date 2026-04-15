// Package builder is the composition root: it wires together infrastructure,
// repositories, and use-cases, then runs the application.
package builder

import (
	"fmt"
	"log"

	"github.com/kainnsoft/duoscans/config"
	"github.com/kainnsoft/duoscans/internal/connection"
	"github.com/kainnsoft/duoscans/internal/ocr"
	"github.com/kainnsoft/duoscans/internal/repository"
	"github.com/kainnsoft/duoscans/internal/usecase"
	commonconfig "github.com/kainnsoft/duoscans/pkg/config"
)

// Run loads configuration and executes the duplicate-detection pipeline.
func Run() {
	cfg, err := commonconfig.FromEnv()
	if err != nil {
		log.Fatalf("reading config: %v", err)
	}

	conn, err := newConnector(&cfg)
	if err != nil {
		log.Fatalf("create connector: %v", err)
	}
	if err := conn.Connect(); err != nil {
		log.Fatalf("connect to device: %v", err)
	}
	defer conn.Close()

	ocrSvc := ocr.New()
	defer ocrSvc.Close()

	deviceRepo := repository.NewDeviceRepository(conn, cfg.TempDir)
	logRepo, err := repository.NewCSVLogRepository(cfg.LogFile)
	if err != nil {
		log.Fatalf("init log repository: %v", err)
	}

	uc := &usecase.FindDuplicatesUseCase{
		Screenshots: deviceRepo,
		OCR:         ocrSvc,
		Log:         logRepo,
	}

	n, err := uc.Execute(usecase.FindDuplicatesInput{
		GalleryPath: cfg.GalleryPath,
		NumFiles:    cfg.NumFiles,
	})
	if err != nil {
		log.Fatalf("execute: %v", err)
	}

	if n > 0 {
		fmt.Printf("results written to %s\n", cfg.LogFile)
	}
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
