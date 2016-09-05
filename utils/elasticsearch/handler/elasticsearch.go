// Package elasticsearch implements an Elasticsearch batch handler.
package elasticsearch

import (
	"fmt"
	"io"
	stdlog "log"
	"os"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/tj/go-elastic/batch"
)

// Elasticsearch interface.
type Elasticsearch interface {
	Bulk(io.Reader) error
}

// Config for handler.
type Config struct {
	BufferSize int           // BufferSize is the number of logs to buffer before flush (default: 100)
	Client     Elasticsearch // Client for ES
	Prefix     string        // Prefix for the index - The index will be prefix-YY-MM-DD (default: logs)
	Hostname   string        // Hostname to add to the logs
}

// defaults applies defaults to the config.
func (c *Config) defaults() {
	if c.BufferSize == 0 {
		c.BufferSize = 100
	}
	if c.Prefix == "" {
		c.Prefix = "logs"
	}
	if c.Hostname == "" {
		c.Hostname, _ = os.Hostname()
	}
}

// Handler implementation.
type Handler struct {
	*Config

	mu    sync.Mutex
	batch *batch.Batch
}

// indexName returns the index for the configured
func (h *Handler) indexName() string {
	return fmt.Sprintf("%s-%s", h.Config.Prefix, time.Now().Format("2006.01.02"))
}

// New handler with BufferSize
func New(config *Config) *Handler {
	config.defaults()
	return &Handler{
		Config: config,
	}
}

// HandleLog implements log.Handler.
func (h *Handler) HandleLog(e *log.Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.batch == nil {
		h.batch = &batch.Batch{
			Elastic: h.Client,
			Index:   h.indexName(),
			Type:    "log",
		}
	}

	// Map fields
	for k, v := range e.Fields {
		switch t := v.(type) {
		case []byte: // Convert []byte to HEX-string
			e.Fields[k] = fmt.Sprintf("%X", t)
		}
	}

	e.Timestamp = e.Timestamp.UTC()

	if h.Hostname != "" {
		e.Fields["hostname"] = h.Hostname
	}

	h.batch.Add(e)

	if h.batch.Size() >= h.BufferSize {
		go h.flush(h.batch)
		h.batch = nil
	}

	return nil
}

// flush the given `batch` asynchronously.
func (h *Handler) flush(batch *batch.Batch) {
	size := batch.Size()
	if err := batch.Flush(); err != nil {
		stdlog.Printf("log/elastic: failed to flush %d logs: %s", size, err)
	}
}
