package connection

import "fmt"

// WiFi implements Connector over a wireless ADB or custom TCP connection.
// Full implementation is a placeholder pending protocol choice.
type WiFi struct {
	// Address is the device IP and port, e.g. "192.168.1.42:5555".
	Address string
}

func (w *WiFi) Connect() error {
	return fmt.Errorf("Wi-Fi connection: not yet implemented (address: %s)", w.Address)
}

func (w *WiFi) ListFiles(remoteDir string, limit int) ([]string, error) {
	return nil, fmt.Errorf("Wi-Fi connection: not yet implemented")
}

func (w *WiFi) Pull(remotePath, localPath string) error {
	return fmt.Errorf("Wi-Fi connection: not yet implemented")
}

func (w *WiFi) Close() error { return nil }
