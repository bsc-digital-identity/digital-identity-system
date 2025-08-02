package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
)

const defaultBase = "http://localhost:9000"

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	base := os.Getenv("API_BASE")
	if base == "" {
		base = defaultBase
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "create":
		msg := fmt.Sprintf("{\"identity_name\":\"%s\"}", mustArg(args, 0))
		do("POST", base+"/api/v1", bytes.NewBufferString(msg))
	case "get":
		do("GET", base+"/api/v1/"+mustArg(args, 0), bytes.NewBufferString(""))
	case "verify":
		msg := "{\"identity_id\":\"" + mustArg(args, 0) + "\",\"zkp_schema\":" + mustArg(args, 1) + "}"
		do("POST", base+"/api/v1/verify", bytes.NewBufferString(msg))
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println(`Usage: cli <command> [options]

Commands:
  create  <name>	         POST /api/v1/
  get     <id>               GET  /api/v1/:id
  verify  <id> <schema>      POST /api/v1/verify

Environment:
  API_BASE   override default http://localhost:9000
`)
}

func mustArg(args []string, idx int) string {
	if len(args) <= idx {
		fmt.Fprintf(os.Stderr, "Missing required argument #%d for command.\n\n", idx+1)
		usage()
		os.Exit(3)
	}
	return args[idx]
}

func do(method, url string, body io.Reader) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create request: %v\n", err)
		os.Exit(10)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Request failed: %v\n", err)
		os.Exit(11)
	}
	defer res.Body.Close()

	fmt.Printf("→ %s %s\n", method, url)
	fmt.Printf("← %d %s\n\n", res.StatusCode, http.StatusText(res.StatusCode))
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		fmt.Fprintf(os.Stderr, "Error: HTTP %d - %s\n", res.StatusCode, http.StatusText(res.StatusCode))
	}
	if _, err := io.Copy(os.Stdout, res.Body); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read response body: %v\n", err)
	}
	fmt.Println()
}
