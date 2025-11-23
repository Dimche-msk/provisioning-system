package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"provisioning-system/internal/broadcaster"
)

type DebugHandler struct {
	Broadcaster *broadcaster.Broadcaster
}

func NewDebugHandler(b *broadcaster.Broadcaster) *DebugHandler {
	return &DebugHandler{
		Broadcaster: b,
	}
}

// StreamLogs handles GET /api/debug/logs (SSE)
func (h *DebugHandler) StreamLogs(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Subscribe to logs
	logChan := h.Broadcaster.Subscribe()
	defer h.Broadcaster.Unsubscribe(logChan)

	// Notify client of connection
	fmt.Fprintf(w, "data: %s\n\n", "connected")
	w.(http.Flusher).Flush()

	// Loop to send events
	for {
		select {
		case event := <-logChan:
			data, err := json.Marshal(event)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			return
		}
	}
}
