package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	"github.com/mitchellh/mapstructure"
)

// ["MaskedEmail/set", idkyet, 0]

const auth = ""
const accID = ""

const apiEndpoint = "https://api.fastmail.com/jmap/api/"
const appName = "myapp"

func main() {
	flag.Parse()

	r := MethodCall{
		MethodName: "MaskedEmail/set",
		Payload:    NewMethodCallCreate(accID, appName, "hoge.com"),
		Payload2:   "0",
	}

	cmer := APIRequest{
		Using: []string{
			"urn:ietf:params:jmap:core",
			"https://www.fastmail.com/dev/maskedemail",
		},
		MethodCalls: []MethodCall{r},
	}

	reqJson, err := json.Marshal(cmer)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%v", string(reqJson))

	req, err := http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(reqJson))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", fmt.Sprintf("Bearer %s", auth))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	buf := &bytes.Buffer{}
	buf.ReadFrom(res.Body)

	var genRe APIResponse
	json.Unmarshal(buf.Bytes(), &genRe)

	// TODO: for debug. remove me.
	func(v interface{}) {
		j, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}
		buf := bytes.NewBuffer(j)
		fmt.Printf("%v\n", buf.String())
	}(genRe.MethodResponsesParsed[0].Payload)

	var pl MethodResponseCreate
	err = mapstructure.Decode(genRe.MethodResponsesParsed[0].Payload, &pl)
	if err != nil {
		panic(err)
	}

	// TODO: for debug. remove me.
	func(v interface{}) {
		j, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}
		buf := bytes.NewBuffer(j)
		fmt.Printf("%v\n", buf.String())
	}(pl)

	r = MethodCall{
		MethodName: "MaskedEmail/set",
		Payload:    NewMethodCallUpdate(accID, pl.Created["myapp"].ID),
		Payload2:   "0",
	}

	cmer = APIRequest{
		Using: []string{
			"urn:ietf:params:jmap:core",
			"https://www.fastmail.com/dev/maskedemail",
		},
		MethodCalls: []MethodCall{r},
	}

	fmt.Println("second res")
	// TODO: for debug. remove me.
	func(v interface{}) {
		j, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}
		buf := bytes.NewBuffer(j)
		fmt.Printf("%v\n", buf.String())
	}(cmer)

	reqJson, err = json.Marshal(cmer)
	if err != nil {
		panic(err)
	}

	req, err = http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(reqJson))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", fmt.Sprintf("Bearer %s", auth))

	res, err = http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	// TODO: for debug. remove me.
	buf = &bytes.Buffer{}
	buf.ReadFrom(res.Body)

	fmt.Printf("%v", buf.String())
}
