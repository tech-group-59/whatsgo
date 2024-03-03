package main

import "C"
import (
	"github.com/otiai10/gosseract/v2"
	"strings"
)

type OCRTracker struct {
	client *gosseract.Client
}

func (tracker *OCRTracker) Init(config *Config) error {
	tracker.client = gosseract.NewClient()
	err := tracker.client.SetLanguage("ukr", "eng", "rus")
	if err != nil {
		return err
	}
	return nil
}

func isImage(file string) bool {

	var extensions = []string{".png", ".jpg", ".jpeg", ".gif", ".bmp", ".jpe", ".webp"}

	for _, ext := range extensions {
		if strings.HasSuffix(file, ext) {
			return true
		}
	}
	return false
}

func (tracker *OCRTracker) TrackMessage(message *TrackableMessage) error {
	var text string

	for _, file := range message.Files {
		if !isImage(file) {
			continue
		}
		log.Infof("Processing image %s", file)
		err := tracker.client.SetImage(file)
		if err != nil {
			return err
		}
		text, _ := tracker.client.Text()
		log.Infof("Extracted text: %s", text)
		message.ParsedContent += text
	}

	message.ParsedContent += text
	return nil
}
