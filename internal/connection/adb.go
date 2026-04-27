package connection

import (
	"fmt"
	"os/exec"
	"strings"
)

// ADB implements Connector using the adb command-line tool.
type ADB struct {
	// Serial is the device serial number. Leave empty to use the only connected device.
	Serial string
}

func (a *ADB) Connect() error {
	args := a.args("devices")
	out, err := exec.Command("adb", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("adb devices: %s: %w", strings.TrimSpace(string(out)), err)
	}

	var hasDevice bool
	var statuses []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "List of devices attached") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		state := fields[1]
		statuses = append(statuses, fmt.Sprintf("%s(%s)", fields[0], state))
		if state == "device" {
			hasDevice = true
		}
	}

	if !hasDevice {
		if len(statuses) == 0 {
			return fmt.Errorf("adb: no devices/emulators found")
		}
		return fmt.Errorf("adb: no ready device, found: %s", strings.Join(statuses, ", "))
	}
	return nil
}

func (a *ADB) ListFiles(remoteDir string, limit int) ([]string, error) {
	args := a.args("shell", "ls", "-1", remoteDir)
	out, err := exec.Command("adb", args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("adb ls %s: %s: %w", remoteDir, strings.TrimSpace(string(out)), err)
	}

	all := strings.Split(strings.TrimSpace(string(out)), "\n")
	if limit > 0 && len(all) > limit {
		all = all[:limit]
	}

	// Prepend directory so callers get full paths.
	for i, name := range all {
		all[i] = remoteDir + "/" + strings.TrimSpace(name)
	}
	return all, nil
}

func (a *ADB) Pull(remotePath, localPath string) error {
	args := a.args("pull", remotePath, localPath)
	if out, err := exec.Command("adb", args...).CombinedOutput(); err != nil {
		return fmt.Errorf("adb pull %s: %s: %w", remotePath, out, err)
	}
	return nil
}

func (a *ADB) Close() error { return nil }

// args prepends -s <serial> when a specific device is targeted.
func (a *ADB) args(sub ...string) []string {
	if a.Serial != "" {
		return append([]string{"-s", a.Serial}, sub...)
	}
	return sub
}
