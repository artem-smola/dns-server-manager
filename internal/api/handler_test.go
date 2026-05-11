package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/artem-smola/dns-server-manager/internal/resolv"
)

type fakeManager struct {
	listResp []string
	listErr  error
	addErr   error
	delErr   error
}

func (m *fakeManager) List() ([]string, error) { return m.listResp, m.listErr }
func (m *fakeManager) Add(string) error        { return m.addErr }
func (m *fakeManager) Remove(string) error     { return m.delErr }

func TestHandleList(t *testing.T) {
	h := NewHandler(&fakeManager{listResp: []string{"1.1.1.1"}}, slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/dns", nil)
	rr := httptest.NewRecorder()
	h.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	var resp struct {
		Servers []string `json:"servers"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Servers) != 1 || resp.Servers[0] != "1.1.1.1" {
		t.Fatalf("unexpected response: %#v", resp)
	}
}

func TestHandleAddInvalidJSON(t *testing.T) {
	h := NewHandler(&fakeManager{}, slog.Default())
	req := httptest.NewRequest(http.MethodPost, "/dns", bytes.NewBufferString("{"))
	rr := httptest.NewRecorder()
	h.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestHandleAddConflict(t *testing.T) {
	h := NewHandler(&fakeManager{addErr: resolv.ErrServerAlreadyExists}, slog.Default())
	body := []byte(`{"server":"8.8.8.8"}`)
	req := httptest.NewRequest(http.MethodPost, "/dns", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rr.Code)
	}
}

func TestHandleDeleteNotFound(t *testing.T) {
	h := NewHandler(&fakeManager{delErr: resolv.ErrServerNotFound}, slog.Default())
	body := []byte(`{"server":"8.8.8.8"}`)
	req := httptest.NewRequest(http.MethodDelete, "/dns", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	h.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestHandleListInternalError(t *testing.T) {
	h := NewHandler(&fakeManager{listErr: errors.New("oops")}, slog.Default())
	req := httptest.NewRequest(http.MethodGet, "/dns", nil)
	rr := httptest.NewRecorder()
	h.Routes().ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}
}
