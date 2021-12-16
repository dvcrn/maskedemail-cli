package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/mitchellh/mapstructure"
)

// ["MaskedEmail/set", idkyet, 0]

const apiEndpoint = "https://api.fastmail.com/jmap/api/"

var flagDomain = flag.String("domain", "", "the domain to create the alias for")
var flagAppname = flag.String("appname", "maskedemail-cli", "the appname to identify the creator")
var flagToken = flag.String("token", "", "the token to authenticate with")
var flagAccountID = flag.String("accountid", "", "fastmail account id")

func init() {
	flag.Parse()

	fmt.Println()

	if *flagToken == "" {
		log.Println("-token flag is not set")
		flag.Usage()
		os.Exit(0)
	}

	if *flagAccountID == "" {
		log.Println("-accountid flag is not set")
		flag.Usage()
		os.Exit(0)
	}

	if *flagDomain == "" {
		log.Println("-domain flag is not set")
		flag.Usage()
		os.Exit(0)
	}

	if len(flag.Args()) != 1 {
		log.Println("no argument given. currently supported: create")
		flag.Usage()
		os.Exit(0)
	}
}

func main() {
	// action := flag.Arg(0)
	fmt.Println(*flagAppname)
	// return

	r := MethodCall{
		MethodName: "MaskedEmail/set",
		Payload:    NewMethodCallCreate(*flagAccountID, *flagAppname, *flagDomain),
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

	fmt.Println("first req")
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

	req, err := http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(reqJson))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", fmt.Sprintf("Bearer %s", *flagToken))

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

	createdItem, err := pl.GetCreatedItem()
	if err != nil {
		panic(err)
	}

	r = MethodCall{
		MethodName: "MaskedEmail/set",
		Payload:    NewMethodCallUpdate(*flagAccountID, createdItem.ID),
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
	req.Header.Set("authorization", fmt.Sprintf("Bearer %s", *flagToken))

	res, err = http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	// TODO: for debug. remove me.
	buf = &bytes.Buffer{}
	buf.ReadFrom(res.Body)

	fmt.Printf("%v", buf.String())
}
