package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type WebhookTracker struct {
	WebhookURLs []string
}

func (w *WebhookTracker) Init(config *Config) error {
	w.WebhookURLs = config.Webhook.URLs
	return nil
}

func (w *WebhookTracker) TrackMessage(message *TrackableMessage) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	var (
		allErrors []error
		mutex     sync.Mutex
		wg        sync.WaitGroup
	)

	for _, url := range w.WebhookURLs {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
			if err != nil {
				mutex.Lock()
				allErrors = append(allErrors, fmt.Errorf("failed to send webhook to %s: %v", url, err))
				mutex.Unlock()
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				mutex.Lock()
				allErrors = append(allErrors, fmt.Errorf("failed to send webhook to %s: %s", url, resp.Status))
				mutex.Unlock()
			}
		}(url)
	}

	wg.Wait()

	if len(allErrors) > 0 {
		return fmt.Errorf("encountered errors: %v", allErrors)
	}

	return nil
}
