package notification

import (
	"fmt"
	"net/http"
	"time"

	common_errors "github.com/astrokiran/nimbus/internal/common/errors"
	"github.com/google/uuid"
)

// StreamNotifications handles server-sent events for notifications.
func (n *Notification) StreamNotifications(w http.ResponseWriter, r *http.Request) {
	// Get parameters from the request
	fmt.Println(r.URL.Query().Get("ID"))
	id, err := uuid.Parse(r.URL.Query().Get("ID"))
	if err != nil {
		common_errors.ErrorMessage(w, r, http.StatusBadRequest, "Invalid ID", nil)
		return
	}
	userType := r.URL.Query().Get("Type")
	fmt.Printf("ID: %s, Type: %s\n", id, userType)

	// Set headers for SSE
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Initialize connections if not already done
	if n.userConnections == nil {
		n.userConnections = make(map[uuid.UUID]http.ResponseWriter)
	}
	if n.consultantConnections == nil {
		n.consultantConnections = make(map[uuid.UUID]http.ResponseWriter)
	}

	// Register the connection based on user type
	n.mu.Lock()

	if userType == "User" {
		n.userConnections[id] = w
	} else if userType == "Consultant" {
		n.consultantConnections[id] = w
	} else {
		http.Error(w, "Invalid user type", http.StatusBadRequest)
		return
	}

	n.mu.Unlock()

	// Create a channel to manage connection termination
	closeChan := make(chan struct{})
	// Continuous streaming of notifications
	go func() {
		for {
			select {
			case <-closeChan:
				return // Exit the goroutine if the channel is closed
			default:

				if w != nil {
					w.(http.Flusher).Flush()
				} else {
					closeChan <- struct{}{}
				}
				time.Sleep(100 * time.Millisecond) // Wait for 100ms before sending the next event
			}
		}
	}()

	// Wait for the connection to be closed
	<-r.Context().Done()
	close(closeChan)
}
