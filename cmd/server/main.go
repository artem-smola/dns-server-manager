package main

import (
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/artem-smola/dns-server-manager/internal/api"
	"github.com/artem-smola/dns-server-manager/internal/resolv"
)

func main() {
	addr := flag.String("addr", ":8080", "server address")
	resolvPath := flag.String("resolv-conf", "/etc/resolv.conf", "path to resolv.conf")
	logPath := flag.String("log-file", "", "optional path to log file")
	flag.Parse()

	logger := buildLogger(*logPath)
	manager := resolv.NewManager(*resolvPath)
	handler := api.NewHandler(manager, logger)

	logger.Info("starting dns-manager server", "addr", *addr, "resolv_conf", *resolvPath)
	if err := http.ListenAndServe(*addr, handler.Routes()); err != nil {
		logger.Error("server stopped", "error", err)
		log.Fatal(err)
	}
}

func buildLogger(logPath string) *slog.Logger {
	if logPath == "" {
		return slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	const rwForOwnerReadOnlyForOthers = 0o644
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, rwForOwnerReadOnlyForOthers)
	if err != nil {
		log.Fatalf("open log file: %v", err)
	}
	return slog.New(slog.NewJSONHandler(f, nil))
}
