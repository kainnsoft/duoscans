package connection

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hanwen/go-mtpfs/mtp"
)

// mtpRoot is the sentinel parent handle for the root of an MTP storage.
const mtpRoot = 0xFFFFFFFF

// USB implements Connector over MTP (Media Transfer Protocol) via libusb.
// DevicePath is an optional regex pattern to select a specific device;
// leave empty to use the first MTP device found.
type USB struct {
	DevicePath string

	dev       *mtp.Device
	storageID uint32
	handles   map[string]uint32 // remotePath → MTP object handle
}

func (u *USB) Connect() error {
	openDevice := func() (*mtp.Device, error) {
		dev, err := mtp.SelectDevice(u.DevicePath)
		if err != nil {
			return nil, fmt.Errorf("mtp: select device: %w", err)
		}
		if err := dev.Configure(); err != nil {
			dev.Done()
			return nil, fmt.Errorf("mtp: configure: %w", err)
		}
		return dev, nil
	}

	dev, err := openDevice()
	if err != nil {
		return err
	}

	// On Android, storage access may appear only after the user confirms
	// the "Allow access to phone data" prompt. Retry briefly to avoid races.
	var storageIDs mtp.Uint32Array
	var lastErr error
	for range 12 {
		storageIDs = mtp.Uint32Array{}
		if err := dev.GetStorageIDs(&storageIDs); err != nil {
			lastErr = err
			// Some phones re-enumerate after permission approval, which
			// invalidates the previous libusb handle. Reopen transparently.
			if strings.Contains(err.Error(), "LIBUSB_ERROR_NO_DEVICE") ||
				strings.Contains(err.Error(), "device is not open") {
				dev.CloseSession()
				dev.Close()
				dev.Done()
				reopened, reopenErr := openDevice()
				if reopenErr != nil {
					lastErr = reopenErr
				} else {
					dev = reopened
				}
			}
		} else {
			lastErr = nil
			if len(storageIDs.Values) > 0 {
				break
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	if lastErr != nil {
		dev.CloseSession()
		dev.Close()
		dev.Done()
		return fmt.Errorf("mtp: get storage IDs: %w", lastErr)
	}
	if len(storageIDs.Values) == 0 {
		dev.CloseSession()
		dev.Close()
		dev.Done()
		return fmt.Errorf("mtp: no storage found on device (did you approve phone data access?)")
	}

	u.dev = dev
	u.storageID = storageIDs.Values[0]
	u.handles = make(map[string]uint32)
	return nil
}

// ListFiles returns up to limit file paths from remoteDir on the device.
func (u *USB) ListFiles(remoteDir string, limit int) ([]string, error) {
	parentHandle, err := u.resolveDir(remoteDir)
	if err != nil {
		return nil, fmt.Errorf("mtp: resolve %s: %w", remoteDir, err)
	}

	var objectHandles mtp.Uint32Array
	if err := u.dev.GetObjectHandles(u.storageID, 0, parentHandle, &objectHandles); err != nil {
		return nil, fmt.Errorf("mtp: list %s: %w", remoteDir, err)
	}

	var paths []string
	for _, h := range objectHandles.Values {
		if limit > 0 && len(paths) >= limit {
			break
		}
		var info mtp.ObjectInfo
		if err := u.dev.GetObjectInfo(h, &info); err != nil {
			continue
		}
		// Skip sub-directories.
		if info.ObjectFormat == mtp.OFC_Association {
			continue
		}
		fullPath := remoteDir + "/" + info.Filename
		u.handles[fullPath] = h
		paths = append(paths, fullPath)
	}
	return paths, nil
}

// Pull downloads the file at remotePath to localPath.
func (u *USB) Pull(remotePath, localPath string) error {
	handle, ok := u.handles[remotePath]
	if !ok {
		return fmt.Errorf("mtp: unknown path %q — call ListFiles first", remotePath)
	}

	f, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("mtp: create local file: %w", err)
	}
	defer f.Close()

	if err := u.dev.GetObject(handle, f); err != nil {
		os.Remove(localPath)
		return fmt.Errorf("mtp: get object %s: %w", remotePath, err)
	}
	return nil
}

func (u *USB) Close() error {
	if u.dev == nil {
		return nil
	}
	u.dev.CloseSession()
	u.dev.Close()
	u.dev.Done()
	u.dev = nil
	return nil
}

// resolveDir walks the MTP object tree to find the handle for a directory path.
func (u *USB) resolveDir(path string) (uint32, error) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	parent := uint32(mtpRoot)

	for _, part := range parts {
		if part == "" {
			continue
		}
		var children mtp.Uint32Array
		if err := u.dev.GetObjectHandles(u.storageID, 0, parent, &children); err != nil {
			return 0, err
		}
		found := false
		for _, h := range children.Values {
			var info mtp.ObjectInfo
			if err := u.dev.GetObjectInfo(h, &info); err != nil {
				continue
			}
			if info.Filename == part {
				parent = h
				found = true
				break
			}
		}
		if !found {
			return 0, fmt.Errorf("directory %q not found", part)
		}
	}
	return parent, nil
}
