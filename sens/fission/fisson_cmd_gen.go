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
			cmd := bytes.NewBufferString("---\nfission fn create --name ")
			fmt.Fprintln(cmd, k, " --env go --src ", f["src"], " --entrypoint ", f["entry"])
			fmt.Fprint(cmd, "fission route create --name ", k, " --method ", f["method"], " --url /", k, " --function ", k, " --createingress")
			fmt.Fprintln(os.Stdout, cmd.String())
		}
	}
}

func main() {
	Generate()
}
