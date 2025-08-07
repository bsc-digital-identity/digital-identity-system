package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const defaultBase = "http://localhost:9000"

func main() {
	base := os.Getenv("API_BASE")
	if base == "" {
		base = defaultBase
	}

	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	createName := createCmd.String("name", "", "Identity name (required)")

	getCmd := flag.NewFlagSet("get", flag.ExitOnError)
	getID := getCmd.String("id", "", "Identity ID (required)")

	verifyCmd := flag.NewFlagSet("verify", flag.ExitOnError)
	verifyID := verifyCmd.String("id", "", "Identity ID (required)")
	verifySchema := verifyCmd.String("schema", "", "ZKP schema (required)")

	if len(os.Args) < 2 {
		usage()
		return
	}

	switch os.Args[1] {
	case "create":
		createCmd.Parse(os.Args[2:])
		if *createName == "" {
			fmt.Fprintln(os.Stderr, "Missing required flag: -name")
			createCmd.Usage()
			os.Exit(1)
		}
		req := struct {
			IdentityName string `json:"identity_name"`
		}{
			IdentityName: *createName,
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
		getCmd.Parse(os.Args[2:])
		if *getID == "" {
			fmt.Fprintln(os.Stderr, "Missing required flag: -id")
			getCmd.Usage()
			os.Exit(1)
		}
		err := do("GET", base+"/api/v1/"+*getID, nil)
		if err != nil {
			log.Fatal(err)
		}

	case "verify":
		verifyCmd.Parse(os.Args[2:])
		if *verifyID == "" || *verifySchema == "" {
			fmt.Fprintln(os.Stderr, "Missing required flags: -id and -schema")
			verifyCmd.Usage()
			os.Exit(1)
		}
		req := struct {
			IdentityName string `json:"identity_name"`
			ZKP_Schema   string `json:"zkp_schema"`
		}{
			IdentityName: *verifyID,
			ZKP_Schema:   *verifySchema,
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
		usage()
		log.Fatalf("Unknown command: %s\n\n", os.Args[1])
	}
}

func usage() {
	fmt.Println(`Usage: cli <command> [flags]

Commands:
  create   -name <name>                 POST /api/v1/
  get      -id <id>                     GET  /api/v1/:id
  verify   -id <id> -schema <schema>    POST /api/v1/verify

Flags:
  -name      Identity name (for create)
  -id        Identity ID (for get/verify)
  -schema    ZKP schema (for verify)

Environment:
  API_BASE   override default http://localhost:9000
`)
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
