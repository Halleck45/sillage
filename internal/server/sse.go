package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Event est un événement diffusé sur le flux SSE /api/events.
type Event struct {
	Name string
	Data any
}

// Hub diffuse des événements à tous les clients SSE abonnés. Un client lent
// n'empêche jamais les autres de recevoir leurs événements (envoi non bloquant).
type Hub struct {
	mu          sync.Mutex
	subscribers map[chan Event]struct{}
}

// NewHub crée un hub SSE prêt à l'emploi.
func NewHub() *Hub {
	return &Hub{subscribers: map[chan Event]struct{}{}}
}

// Subscribe enregistre un nouveau client et retourne son canal ainsi qu'une
// fonction de désinscription à appeler impérativement quand le client se déconnecte.
func (h *Hub) Subscribe() (chan Event, func()) {
	ch := make(chan Event, 64)
	h.mu.Lock()
	h.subscribers[ch] = struct{}{}
	h.mu.Unlock()
	unsub := func() {
		h.mu.Lock()
		delete(h.subscribers, ch)
		h.mu.Unlock()
	}
	return ch, unsub
}

// Publish envoie un événement à tous les abonnés. Si le buffer d'un client
// est plein, l'événement est simplement abandonné pour ce client (non bloquant).
func (h *Hub) Publish(ev Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.subscribers {
		select {
		case ch <- ev:
		default:
		}
	}
}

// ServeSSE est le handler HTTP du flux /api/events.
func (h *Hub) ServeSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "flux non supporté", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	ch, unsub := h.Subscribe()
	defer unsub()

	ping := time.NewTicker(25 * time.Second)
	defer ping.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case ev := <-ch:
			data, err := json.Marshal(ev.Data)
			if err != nil {
				continue
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Name, data)
			flusher.Flush()
		case <-ping.C:
			fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		}
	}
}
