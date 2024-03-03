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
The default configuration file is `config.yaml` and it should be placed in the same directory as the binary.

```yaml
# config.yaml
