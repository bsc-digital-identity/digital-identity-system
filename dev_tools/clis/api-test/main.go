package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
		name, err := mustArg(args, 0)
		if err != nil {
			usage()
			log.Fatal(err)
		}
		req := struct {
			IdentityName string `json:"identity_name"`
		}{
			IdentityName: name,
		}
		msg, err := json.Marshal(req)
		if err != nil {
			log.Fatal(err)
		}
		err = do("POST", base+"/api/v1", bytes.NewBuffer(msg))
		if err != nil {
			log.Fatal(err)
		}

	case "get":
		name, err := mustArg(args, 0)
		if err != nil {
			usage()
			log.Fatal(err)
		}
		err = do("GET", base+"/api/v1/"+name, bytes.NewBufferString(""))
		if err != nil {
			log.Fatal(err)
		}

	case "verify":
		name, err := mustArg(args, 0)
		if err != nil {
			usage()
			log.Fatal(err)
		}
		schema, err := mustArg(args, 1)
		if err != nil {
			usage()
			log.Fatal(err)
		}

		req := struct {
			IdentityName string `json:"identity_name"`
			ZKP_Schema   string `json:"zkp_schema"`
		}{
			IdentityName: name,
			ZKP_Schema:   schema,
		}
		msg, err := json.Marshal(req)
		if err != nil {
			log.Fatal(err)
		}

		err = do("POST", base+"/api/v1/verify", bytes.NewBuffer(msg))
		if err != nil {
			log.Fatal(err)
		}

	default:
		log.Fatalf("Unknown command: %s\n\n", cmd)
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
  API_BASE   override default http://localhost:9000`)
}

func mustArg(args []string, idx int) (string, error) {
	if len(args) <= idx {
		return "", fmt.Errorf("Missing required argument #%d for command.\n\n", idx+1)
	}
	return args[idx], nil
}

func do(method, url string, body io.Reader) error {
	var req *http.Request
	var err error
	if method == "GET" {
		req, err = http.NewRequest(method, url, nil)
	} else {
		req, err = http.NewRequest(method, url, body)
	}

	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
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
	return err
}
