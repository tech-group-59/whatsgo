# What's Go

This is a Go-based project that interacts with WhatsApp's Web API. 
It uses the [`whatsmeow`](https://github.com/tulir/whatsmeow/tree/main) library to provide a
command-line interface for WhatsApp.

## Features

- Send and receive messages
- Manage group chats
- Handle media messages
- Interact with WhatsApp's Web API

## Requirements

- Go 1.22
- SQLite3
- QR Terminal

## Dependencies

- github.com/mattn/go-sqlite3
- github.com/mdp/qrterminal/v3
- go.mau.fi/whatsmeow

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
Usage of ./bin/whatsgo-arm64-mac:
  -db-address string
    	Database address (default "file:whatsgo.db?_foreign_keys=on")
  -db-dialect string
    	Database dialect (sqlite3 or postgres) (default "sqlite3")
  -debug
    	Enable debug logs?
  -files-folder string
    	Folder to save files to (default "files")
  -request-full-sync
    	Request full (1 year) history sync when logging in?
```