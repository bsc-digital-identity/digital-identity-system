package main

import (
	"bytes"
	"flag"
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
		// postJSON(base+"/api/v1/identity", args)
		msg := fmt.Sprintf("{\"identity_name\":\"%s\"}", mustArg(args, 0))
		r := bytes.NewBufferString(msg)
		do("POST", base+"/api/v1/identity", r)
	case "get":
		getJSON(base+"/api/v1/identity", args)
	case "verify":
		postJSON(base+"/api/v1/identity/verify", args)
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Println(`Usage: cli <command> [options]

Commands:
  create   -d '{"identity_name":"foo"}'   POST /api/v1/identity
  get      <id>                            GET  /api/v1/identity/:id
  verify   -d '{"identity_id":"id", "schema":"bar"}' POST /api/v1/identity/verify

Environment:
  API_BASE   override default http://localhost:8080
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

func get(url string) {
	do("GET", url, nil)
}

func del(url string) {
	do("DELETE", url, nil)
}

func postNoBody(url string) {
	do("POST", url, nil)
}

func getJSON(url string, args []string) {
	// data := pickJSON(args)
	// do("GET", url, data)
	msg := fmt.Sprintf("{\"identity_name\":\"%s\"}", mustArg(args, 0))
	r := bytes.NewBufferString(msg)
	do("GET", url, r)
}

func postJSON(url string, args []string) {
	data := pickJSON(args)
	do("POST", url, data)
}

func putJSON(url string, args []string) {
	data := pickJSON(args)
	do("PUT", url, data)
}

func pickJSON(args []string) io.Reader {
	fs := flag.NewFlagSet("", flag.ExitOnError)
	body := fs.String("d", "", "request JSON body")
	fs.Parse(args)
	var r io.Reader
	if *body != "" {
		r = bytes.NewBufferString(*body)
	} else {
		// read from stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			r = os.Stdin
		}
	}
	return r
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
