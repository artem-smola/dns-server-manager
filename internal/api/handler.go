package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/artem-smola/dns-server-manager/internal/resolv"
)

type DNSManager interface {
	List() ([]string, error)
	Add(server string) error
	Remove(server string) error
}

type Handler struct {
	manager DNSManager
	log     *slog.Logger
}

func NewHandler(manager DNSManager, log *slog.Logger) *Handler {
	return &Handler{
		manager: manager,
		log:     log,
	}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/dns", h.handleDNS)
	mux.HandleFunc("/healthz", h.handleHealth)
	return loggingMiddleware(h.log, mux)
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleDNS(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleList(w)
	case http.MethodPost:
		h.handleAdd(w, r)
	case http.MethodDelete:
		h.handleRemove(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

type listResponse struct {
	Servers []string `json:"servers"`
}

func (h *Handler) handleList(w http.ResponseWriter) {
	servers, err := h.manager.List()
	if err != nil {
		h.log.Error("failed to list dns-servers", "error", err)
		writeError(w, http.StatusInternalServerError, "failed to read dns-servers")
		return
	}
	writeJSON(w, http.StatusOK, listResponse{Servers: servers})
}

type changeRequest struct {
	Server string `json:"server"`
}

func (h *Handler) handleAdd(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeChangeRequest(w, r)
	if !ok {
		return
	}

	err := h.manager.Add(req.Server)
	if err != nil {
		switch {
		case errors.Is(err, resolv.ErrInvalidServer):
			writeError(w, http.StatusBadRequest, "invalid dns-server")
		case errors.Is(err, resolv.ErrServerAlreadyExists):
			writeError(w, http.StatusConflict, "dns-server already exists")
		default:
			h.log.Error("failed to add dns-server", "server", req.Server, "error", err)
			writeError(w, http.StatusInternalServerError, "failed to add dns-server")
		}
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "created"})
}

func (h *Handler) handleRemove(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeChangeRequest(w, r)
	if !ok {
		return
	}

	err := h.manager.Remove(req.Server)
	if err != nil {
		switch {
		case errors.Is(err, resolv.ErrInvalidServer):
			writeError(w, http.StatusBadRequest, "invalid dns-server")
		case errors.Is(err, resolv.ErrServerNotFound):
			writeError(w, http.StatusNotFound, "dns-server not found")
		default:
			h.log.Error("failed to remove dns-server", "server", req.Server, "error", err)
			writeError(w, http.StatusInternalServerError, "failed to remove dns-server")
		}
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "removed"})
}

func decodeChangeRequest(w http.ResponseWriter, r *http.Request) (changeRequest, bool) {
	var req changeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return changeRequest{}, false
	}
	if req.Server == "" {
		writeError(w, http.StatusBadRequest, "field 'server' is required")
		return changeRequest{}, false
	}
	return req, true
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func loggingMiddleware(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)
		log.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", time.Since(start).String(),
			"remote_addr", r.RemoteAddr,
		)
	})
}
