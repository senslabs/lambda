package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func Generate() {
	if f, err := os.Open("fisson_input.json"); err != nil {
		log.Fatal(err)
	} else {
		var m map[string]interface{}
		decoder := json.NewDecoder(f)
		if err := decoder.Decode(&m); err != nil {
			log.Fatal(err)
		}
		for k, v := range m {
			f := v.(map[string]interface{})
			cmd := bytes.NewBufferString("#---")
			fmt.Fprintln(cmd, "\nfission fn delete --name ", k)
			fmt.Fprintln(cmd, "fission fn create --name ", k, " --env go --src ", f["src"], " --entrypoint ", f["entry"])
			fmt.Fprint(cmd, "fission route create --name ", k, " --method ", f["method"], " --url ", f["path"], " --function ", k)
			fmt.Fprintln(os.Stdout, cmd.String())
		}
	}
}

func GenerateUpdate() {
	if f, err := os.Open("fisson_input.json"); err != nil {
		log.Fatal(err)
	} else {
		var m map[string]interface{}
		decoder := json.NewDecoder(f)
		if err := decoder.Decode(&m); err != nil {
			log.Fatal(err)
		}
		for k, v := range m {
			f := v.(map[string]interface{})
			cmd := bytes.NewBufferString("#---\n")
			fmt.Fprint(cmd, "fission fn update --name ", k, " --env go --src ", f["src"], " --entrypoint ", f["entry"])
			fmt.Fprintln(os.Stdout, cmd.String())
		}
	}
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Pass an argument. create or update")
	} else if os.Args[1] == "create" {
		Generate()
	} else if os.Args[1] == "update" {
		GenerateUpdate()
	}
}
