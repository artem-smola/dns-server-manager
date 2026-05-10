package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	baseURL := flag.String("server", "http://127.0.0.1:8080", "dns-manager server base URL")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  dns-client [--server <URL>] list\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  dns-client [--server <URL>] add <ip>\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  dns-client [--server <URL>] remove <ip>\n")
		fmt.Fprintf(flag.CommandLine.Output(), "\nOptions:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	switch args[0] {
	case "list":
		if len(args) != 1 {
			exitWithMessage("list command does not accept extra arguments")
		}
		if err := listServers(client, *baseURL); err != nil {
			exitWithMessage(err.Error())
		}
	case "add":
		if len(args) != 2 {
			exitWithMessage("expected: add <ip>")
		}
		if err := changeServer(client, *baseURL, http.MethodPost, args[1]); err != nil {
			exitWithMessage(err.Error())
		}
		fmt.Println("server added")
	case "remove":
		if len(args) != 2 {
			exitWithMessage("expected: remove <ip>")
		}
		if err := changeServer(client, *baseURL, http.MethodDelete, args[1]); err != nil {
			exitWithMessage(err.Error())
		}
		fmt.Println("server removed")
	default:
		exitWithMessage("unknown command: " + args[0])
	}
}

func listServers(client *http.Client, baseURL string) error {
	req, err := http.NewRequest(http.MethodGet, strings.TrimRight(baseURL, "/")+"/dns", nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}

	var payload struct {
		Servers []string `json:"servers"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	if len(payload.Servers) == 0 {
		fmt.Println("no DNS servers configured")
		return nil
	}

	for _, s := range payload.Servers {
		fmt.Println(s)
	}
	return nil
}

func changeServer(client *http.Client, baseURL, method, server string) error {
	payload, err := json.Marshal(map[string]string{"server": server})
	if err != nil {
		return fmt.Errorf("prepare request body: %w", err)
	}

	req, err := http.NewRequest(method, strings.TrimRight(baseURL, "/")+"/dns", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 300 {
		return fmt.Errorf("server returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func exitWithMessage(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	fmt.Fprintln(os.Stderr, "run with --help for usage")
	os.Exit(2)
}
