package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type AwsCommand struct {
	Secret  string `json:"secret"`
	Key     string `json:"key"`
	Command string `json"command"`
}

func (ac *AwsCommand) Execute() ([]byte, error) {
	parts := strings.Split(ac.Command, " ")
	args := parts[1:len(parts)]

	if parts[0] != "aws" {
		return nil, errors.New(fmt.Sprintf("Unknown command: %s. Try 'aws help'", parts[0]))
	}

	// add required endpoint to our service to every call
	args = append(args, "--endpoint-url=https://s3.bluearchivedev.com")

	cmd := exec.Command("aws", args...)

	// set up proper environment variables for authentication
	env := os.Environ()
	env = append(env, fmt.Sprintf("AWS_ACCESS_KEY_ID=%s", ac.Key))
	env = append(env, fmt.Sprintf("AWS_SECRET_ACCESS_KEY=%s", ac.Secret))
	env = append(env, "AWS_DEFAULT_REGION=us-east-1")
	cmd.Env = env

	return cmd.CombinedOutput()
}

func renderJson(w http.ResponseWriter, status int, data interface{}) {
	jsonData, err := json.Marshal(data)

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(jsonData)
}

func parseRequest(r *http.Request, data interface{}) error {
	const maxRequestLen = 16 * 1024 * 1024
	lr := io.LimitReader(r.Body, maxRequestLen)
	return json.NewDecoder(lr).Decode(data)
}

func main() {
	http.HandleFunc("/aws", func(w http.ResponseWriter, r *http.Request) {
		type Response struct {
			Result []byte `json:"result"`
		}

		if r.Method == "POST" {
			cmd := &AwsCommand{}
			parseRequest(r, cmd)
			fmt.Printf("Parsed request: %v\n", cmd)

			out, err := cmd.Execute()
			if err != nil {
				http.Error(w, err.Error(), 400)
				return
			}

			resp := &Response{out}
			fmt.Printf("Response will be: %s\n", resp)
			renderJson(w, 200, resp)
		} else {
			resp := &Response{[]byte("Need to post command")}
			renderJson(w, 200, resp)
		}
	})

	log.Fatal(http.ListenAndServe(":4000", nil))
}
