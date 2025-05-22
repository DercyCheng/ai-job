package logger

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type LokiHook struct {
	client    *http.Client
	url       string
	labels    map[string]string
	batch     []logEntry
	batchSize int
}

type logEntry struct {
	Timestamp time.Time
	Labels    map[string]string
	Message   string
	Level     logrus.Level
}

func NewLokiHook(url string, labels map[string]string, batchSize int) *LokiHook {
	return &LokiHook{
		client:    &http.Client{Timeout: 10 * time.Second},
		url:       url,
		labels:    labels,
		batchSize: batchSize,
	}
}

func (h *LokiHook) Fire(entry *logrus.Entry) error {
	labels := make(map[string]string, len(h.labels))
	for k, v := range h.labels {
		labels[k] = v
	}
	for k, v := range entry.Data {
		labels[k] = toString(v)
	}

	h.batch = append(h.batch, logEntry{
		Timestamp: entry.Time,
		Labels:    labels,
		Message:   entry.Message,
		Level:     entry.Level,
	})

	if len(h.batch) >= h.batchSize {
		return h.sendBatch()
	}
	return nil
}

func (h *LokiHook) sendBatch() error {
	if len(h.batch) == 0 {
		return nil
	}

	var streams []map[string]interface{}
	stream := make(map[string]interface{})
	stream["stream"] = h.labels

	var values [][]string
	for _, entry := range h.batch {
		values = append(values, []string{
			entry.Timestamp.Format(time.RFC3339Nano),
			entry.Message,
		})
	}

	stream["values"] = values
	streams = append(streams, stream)

	payload := map[string]interface{}{
		"streams": streams,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", h.url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return err
	}

	h.batch = nil
	return nil
}

func (h *LokiHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func toString(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case error:
		return v.Error()
	default:
		return ""
	}
}
