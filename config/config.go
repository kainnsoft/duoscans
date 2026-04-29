package config

import (
	"fmt"
	"os"
)

// ConnectionType specifies how to connect to the device.
type ConnectionType string

const (
	ConnectionADB  ConnectionType = "adb"
	ConnectionUSB  ConnectionType = "usb"
	ConnectionWiFi ConnectionType = "wifi"
)

// Config holds all application settings.
type Config struct {
	// Connection settings
	ConnectionType ConnectionType `yaml:"connection_type"` // "adb", "usb", or "wifi"
	DeviceAddress  string         `yaml:"device_address"`  // IP:port for Wi-Fi, serial for ADB/USB

	// Download settings
	NumFiles    int    `yaml:"num_files"`    // number of screenshots to download per run
	GalleryPath string `yaml:"gallery_path"` // path on device where screenshots are stored
	TempDir     string `yaml:"temp_dir"`     // local temp directory; empty means system temp

	// Output settings
	LogFile string `yaml:"log_file"` // path to the CSV output log
}

func (c *Config) normalize() {
	if c.TempDir == "" {
		c.TempDir = os.TempDir()
	}
}

// ValidateDownload checks fields required by the download command.
func (c *Config) ValidateDownload() error {
	c.normalize()

	switch c.ConnectionType {
	case ConnectionADB, ConnectionUSB, ConnectionWiFi:
	case "":
		return fmt.Errorf("connection_type is required")
	default:
		return fmt.Errorf("unknown connection_type %q: must be adb, usb, or wifi", c.ConnectionType)
	}

	if c.GalleryPath == "" {
		return fmt.Errorf("gallery_path is required")
	}
	if c.NumFiles <= 0 {
		return fmt.Errorf("num_files must be greater than 0")
	}

	return nil
}

// ValidateFindDuplicates checks fields required by the find-duplicates command.
func (c *Config) ValidateFindDuplicates() error {
	c.normalize()

	if c.LogFile == "" {
		return fmt.Errorf("log_file is required")
	}
	if c.NumFiles <= 0 {
		return fmt.Errorf("num_files must be greater than 0")
	}

	return nil
}
