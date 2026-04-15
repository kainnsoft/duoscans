# duoscans

Connects to an Android smartphone, downloads Duolingo screenshots from the Gallery, extracts English sentences via OCR, and logs duplicates to a CSV file. No files are ever deleted from the device.

---

## Prerequisites

### 1. System libraries

```bash
# OCR engine
sudo apt-get install -y tesseract-ocr libtesseract-dev libleptonica-dev

# USB/MTP support (required even for ADB-only use)
sudo apt-get install -y libusb-1.0-0-dev
```

### 2. ADB (if using `connection_type: adb`)

```bash
sudo apt-get install -y adb
```

Enable **USB Debugging** on the phone:
> Settings → Developer Options → USB Debugging → ON

Connect the phone and confirm it is visible:
```bash
adb devices
```

### 3. MTP (if using `connection_type: usb`)

On the phone, when prompted after plugging in via USB, select:
> **File Transfer / MTP** mode

No extra tools needed — the app talks to the device directly via libusb.

### 4. Wi-Fi (not yet implemented)

`connection_type: wifi` is a stub and will return an error on connect.

---

## Configuration

Set the path to your config file via the `APP_CONF_PATH` environment variable:

```bash
export APP_CONF_PATH=vars/env.local.yaml
```

Edit `vars/env.local.yaml` to match your setup:

| Field | Description |
|---|---|
| `connection_type` | `adb`, `usb`, or `wifi` |
| `device_address` | ADB serial number or MTP device regex (leave empty for first device) |
| `gallery_path` | Path to screenshots folder on the device |
| `num_files` | How many screenshots to download per run |
| `temp_dir` | Local staging folder (empty = system temp) |
| `log_file` | Output CSV path |

---

## Build & Run

```bash
# Build
go build ./cmd/duoscans

# Run
APP_CONF_PATH=vars/env.local.yaml ./duoscans

# Or without building
APP_CONF_PATH=vars/env.local.yaml go run ./cmd/duoscans
```

---

## Output

Duplicates are appended to the CSV file defined in `log_file`:

```
timestamp,sentence,file_count,local_path,remote_path
2026-04-14T10:00:00Z,"translate this sentence",2,...,...
```

The file grows across runs — each run appends new rows with a timestamp.
