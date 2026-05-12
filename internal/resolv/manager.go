package resolv

import (
	"errors"
	"fmt"
	"net/netip"
	"os"
	"slices"
	"strings"
	"sync"
)

var (
	ErrInvalidServer       = errors.New("invalid dns-server")
	ErrServerAlreadyExists = errors.New("dns-server already exists")
	ErrServerNotFound      = errors.New("dns-server not found")
)

type Manager struct {
	path string
	mtx  sync.Mutex
}

func NewManager(path string) *Manager {
	return &Manager{path: path}
}

func (m *Manager) List() ([]string, error) {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	doc, err := readDocument(m.path)
	if err != nil {
		return nil, err
	}

	out := make([]string, len(doc.servers))
	copy(out, doc.servers)
	return out, nil
}

func (m *Manager) Add(server string) error {
	if !isValidIP(server) {
		return ErrInvalidServer
	}

	m.mtx.Lock()
	defer m.mtx.Unlock()

	doc, err := readDocument(m.path)
	if err != nil {
		return err
	}

	if slices.Contains(doc.servers, server) {
		return ErrServerAlreadyExists
	}
	doc.servers = append(doc.servers, server)
	return writeDocument(m.path, doc)
}

func (m *Manager) Remove(server string) error {
	if !isValidIP(server) {
		return ErrInvalidServer
	}

	m.mtx.Lock()
	defer m.mtx.Unlock()

	doc, err := readDocument(m.path)
	if err != nil {
		return err
	}

	ind := slices.Index(doc.servers, server)
	if ind == -1 {
		return ErrServerNotFound
	}

	doc.servers = slices.Delete(doc.servers, ind, ind+1)
	return writeDocument(m.path, doc)
}

type document struct {
	preamble []string
	servers  []string
}

func readDocument(path string) (document, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return document{}, nil
		}
		return document{}, fmt.Errorf("read document resolv.conf: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	doc := document{
		preamble: make([]string, 0, len(lines)),
		servers:  make([]string, 0),
	}

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "" {
			if line != "" {
				doc.preamble = append(doc.preamble, line)
			}
			continue
		}

		fields := strings.Fields(trimmedLine)
		if len(fields) >= 2 && fields[0] == "nameserver" && isValidIP(fields[1]) {
			doc.servers = append(doc.servers, fields[1])
			continue
		}

		doc.preamble = append(doc.preamble, line)
	}

	return doc, nil
}

func writeDocument(path string, doc document) error {
	var b strings.Builder
	for _, line := range doc.preamble {
		if strings.TrimSpace(line) == "" {
			b.WriteString("\n")
			continue
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	for _, s := range doc.servers {
		b.WriteString("nameserver ")
		b.WriteString(s)
		b.WriteString("\n")
	}

	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return fmt.Errorf("write document resolv.conf: %w", err)
	}
	return nil
}

func isValidIP(candidate string) bool {
	_, err := netip.ParseAddr(candidate)
	return err == nil
}
