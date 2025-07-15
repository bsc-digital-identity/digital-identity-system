package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
)

const defaultBase = "http://localhost:8080"

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
	// Identity
	case "identity-create":
		postJSON(base+"/identity/create", args)
	case "identity-verify":
		postJSON(base+"/identity/verify", args)
	case "identity-recover":
		postJSON(base+"/identity/recover", args)
	case "identity-get":
		get(base+"/identity/"+mustArg(args, 0))
	case "identity-list":
		get(base + "/identity")
	case "identity-update":
		putJSON(base+"/identity/"+mustArg(args, 0), args)
	case "identity-delete":
		del(base + "/identity/" + mustArg(args, 0))
	case "me":
		get(base + "/me")
	case "me-creds":
		get(base + "/me/credentials")

	// Credential
	case "cred-issue":
		postJSON(base+"/credential/issue", args)
	case "cred-verify":
		postJSON(base+"/credential/verify", args)
	case "cred-request":
		postJSON(base+"/credential/request", args)
	case "cred-get":
		get(base+"/credential/"+mustArg(args, 0))
	case "cred-list":
		get(base + "/credentials")
	case "cred-delete":
		del(base + "/credential/" + mustArg(args, 0))

	// Session & Admin & Health
	case "logout":
		postNoBody(base + "/logout")
	case "session":
		get(base + "/session")
	case "stats":
		get(base + "/admin/stats")
	case "health":
		get(base + "/healthz")

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`Usage: cli <command> [options]

Commands:
  identity-create   -d '{"foo":"bar"}'      POST /identity/create
  identity-verify   -d '{"foo":"bar"}'      POST /identity/verify
  identity-recover  -d '{"foo":"bar"}'      POST /identity/recover
  identity-get      <id>                    GET  /identity/:id
  identity-list                            GET  /identity
  identity-update   <id> -d '{"foo":"bar"}' PUT  /identity/:id
  identity-delete   <id>                    DELETE /identity/:id
  me                                      GET  /me
  me-creds                                GET  /me/credentials

  cred-issue        -d '{"foo":"bar"}'      POST /credential/issue
  cred-verify       -d '{"foo":"bar"}'      POST /credential/verify
  cred-request      -d '{"foo":"bar"}'      POST /credential/request
  cred-get          <id>                    GET  /credential/:id
  cred-list                                GET  /credentials
  cred-delete       <id>                    DELETE /credential/:id

  logout                                  POST /logout
  session                                 GET  /session
  stats                                   GET  /admin/stats
  health                                  GET  /healthz

Environment:
  API_BASE   override default http://localhost:8080
`)
}

func mustArg(args []string, idx int) string {
	if len(args) <= idx {
		fmt.Fprintf(os.Stderr, "missing argument %d\n", idx+1)
		usage()
		os.Exit(1)
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
		fmt.Println("req:", err)
		os.Exit(1)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("do:", err)
		os.Exit(1)
	}
	defer res.Body.Close()

	fmt.Printf("→ %s %s\n", method, url)
	fmt.Printf("← %d %s\n\n", res.StatusCode, http.StatusText(res.StatusCode))
	io.Copy(os.Stdout, res.Body)
	fmt.Println()
}

