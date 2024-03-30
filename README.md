# What's Go

This project is a Go-based application that interacts with WhatsApp's Web API
to track messages and files from selected chats.
It uses the [`whatsmeow`](https://github.com/tulir/whatsmeow/tree/main) library
to provide a command-line interface for WhatsApp.

## Features

- Send and receive messages
- Store messages and files in a database
- Store messages and files in Google Drive
- Process images by OCR

## Production Requirements

- Docker

## Usage

### Run in docker

```bash
cp config/demo-config.yaml config/config.yaml
# update config/config.yaml with your settings
mkdir -p data/db
mkdir -p data/files
docker compose run app
```

## Configuration

The application uses a YAML configuration file to store the user's credentials and settings.
The default configuration file is `config.yaml` and it should be placed in the same directory as the binary
or specified using the `--config` flag.

```yaml
chats:
  - id: <chat1-to-track>@s.whatsapp.net
    alias: 'Chat1 alias'
  - id: <chat2-to-track>@s.whatsapp.net
    alias: 'Chat2 alias'
  - id: <chat3-to-track>@s.whatsapp.net
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

## CLI

To get list of groups and contacts enter `listgroups` command.

```bash 

## Trackers

The application uses a tracker to track messages and files from selected chats.
To filter chats that you want to track please use `chats` section at `config.yaml`.
By default, it stores files all images attached to messages.
To configure local path for it use `file_storage_path` section at `config.yaml`.

### DB Tracker

The DB tracker stores messages and files in database tables.
For now, it only supports SQLite3.
It uses the same db that `whatmeow` uses, so it will create a new tables:

- `messages`
- `files`

### Google Drive Tracker

The Google Drive tracker stores messages and files in Google Drive.
It uses the Google Drive API to create a folder for each chat and store the messages and files in it.
To enable it:

- set `true` at `google_cloud.enabled` in `config.yaml`
- create `credentials.json` and set path to it at `google_cloud.credentials_file`.
  Check instruction [here](https://developers.google.com/sheets/api/quickstart/go)
- create folder on Google Drive and set its ID at `google_cloud.folder_id`

### OCR

The OCR tracker uses Tesseract OCR to extract text from images.
To enable it set `true` at `ocr.enabled` in `config.yaml`

## Development

### Requirements

- Go 1.22
- WhatsApp account
- SQLite3
- QR Terminal
- Google Cloud credentials (if using Google Drive)
- Tesseract OCR (if using OCR)

### Building

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
