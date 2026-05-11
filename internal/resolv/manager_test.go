package resolv

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestManagerAddListRemove(t *testing.T) {
	path := filepath.Join(t.TempDir(), "resolv.conf")
	initContent := "#test data\nsearch local\nnameserver 1.1.1.1\n"
	if err := os.WriteFile(path, []byte(initContent), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	manager := NewManager(path)
	if err := manager.Add("8.8.8.8"); err != nil {
		t.Fatalf("add dns-server: %v", err)
	}

	servers, err := manager.List()
	if err != nil {
		t.Fatalf("list dns-servers: %v", err)
	}
	if len(servers) != 2 || servers[0] != "1.1.1.1" || servers[1] != "8.8.8.8" {
		t.Fatalf("unexpected list dns-servers result: %#v", servers)
	}

	if err := manager.Remove("1.1.1.1"); err != nil {
		t.Fatalf("remove dns-server: %v", err)
	}

	servers, err = manager.List()
	if err != nil {
		t.Fatalf("list servers: %v", err)
	}
	if len(servers) != 1 || servers[0] != "8.8.8.8" {
		t.Fatalf("unexpected list dns-servers result after remove: %#v", servers)
	}
}

func TestManagerAddDuplicate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "resolv.conf")
	if err := os.WriteFile(path, []byte("nameserver 9.9.9.9\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	manager := NewManager(path)
	err := manager.Add("9.9.9.9")
	if !errors.Is(err, ErrServerAlreadyExists) {
		t.Fatalf("expected ErrServerExists, got %v", err)
	}
}

func TestManagerRemoveMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "resolv.conf")
	if err := os.WriteFile(path, []byte("nameserver 9.9.9.9\n"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	manager := NewManager(path)
	err := manager.Remove("8.8.8.8")
	if !errors.Is(err, ErrServerNotFound) {
		t.Fatalf("expected ErrServerMissing, got %v", err)
	}
}

func TestManagerRejectInvalidIP(t *testing.T) {
	path := filepath.Join(t.TempDir(), "resolv.conf")
	manager := NewManager(path)

	if err := manager.Add("not-an-ip"); !errors.Is(err, ErrInvalidServer) {
		t.Fatalf("expected ErrInvalidServer from add, got %v", err)
	}
	if err := manager.Remove("not-an-ip"); !errors.Is(err, ErrInvalidServer) {
		t.Fatalf("expected ErrInvalidServer from remove, got %v", err)
	}
}
