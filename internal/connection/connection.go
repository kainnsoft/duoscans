// Package connection defines how the application connects to a smartphone
// and lists files available for download.
package connection

// Connector is the interface every connection type must implement.
type Connector interface {
	// Connect establishes a connection to the device.
	Connect() error

	// ListFiles returns up to limit file paths from the given remote directory.
	ListFiles(remoteDir string, limit int) ([]string, error)

	// Pull downloads a single file from remotePath to localPath.
	Pull(remotePath, localPath string) error

	// Close releases any resources held by the connection.
	Close() error
}
