# What's Go

This project is a Go-based application that interacts with WhatsApp's Web API
to track messages and files from selected chats.
It uses the [`whatsmeow`](https://github.com/tulir/whatsmeow/tree/main) library
to provide a command-line interface for WhatsApp.

## Features

- Send and receive messages
- Store messages and files in a database
- Store messages and files in Google Drive

## Requirements

- Go 1.22
- WhatsApp account
- SQLite3
- QR Terminal
- Google Cloud credentials (if using Google Drive)
- Tesseract OCR (if using OCR)

## Building

This project uses a Makefile for building. You can build for Linux, macOS, and Windows using the following commands:

```bash
make build-linux
make build-mac
make build-windows
```

Or build for all platforms at once:

```bash
make build-all
```

## Usage

After building, you can run the binary for your platform.
The application will provide a command-line interface for interacting with WhatsApp.

```bash
Usage of whatsgo:
  -config string
    	Path to config file (default "config.yaml")
  -debug
    	Enable debug logs?
  -request-full-sync
    	Request full (1 year) history sync when logging in?
```

## Configuration

The application uses a YAML configuration file to store the user's credentials and settings.
The default configuration file is `config.yaml` and it should be placed in the same directory as the binary
or specified using the `--config` flag.

```yaml
chats:
  - <chat-to-track>@s.whatsapp.net
file_storage_path: "folder-to-store-files"
ocr:
  enabled: false
database:
  connection_string: "file:whatsgo.db?_foreign_keys=on"
  dialect: "sqlite3"
google_cloud:
  enabled: true
  credentials_file: "credentials.json" # path to the Google cloud credentials file, details on how to get it here: https://developers.google.com/sheets/api/quickstart/go
  folder_id: "<google-folder-id>"
```

## Trackers

The application uses a tracker to track messages and files from selected chats.

### DB Tracker

The DB tracker stores messages and files in database tables.
For now, it only supports SQLite3.
It uses the same db that `whatmeow` uses, so it will create a new tables:

- `messages`
- `files`


### Google Drive Tracker

The Google Drive tracker stores messages and files in Google Drive.
It uses the Google Drive API to create a folder for each chat and store the messages and files in it.

### OCR

The OCR tracker uses Tesseract OCR to extract text from images.
It uses the `tesseract` command-line tool to check.
Details: 
