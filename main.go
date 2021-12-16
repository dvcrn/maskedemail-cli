package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
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

	client := NewClient(*flagAccountID, *flagToken, *flagAppname)

	res, err := client.CreateMaskedEmail(*flagDomain)
	if err != nil {
		panic(err)
	}

	r := MethodCall{
		MethodName: "MaskedEmail/set",
		Payload:    NewMethodCallUpdate(*flagAccountID, res.ID),
		Payload2:   "0",
	}

	cmer := APIRequest{
		Using: []string{
			"urn:ietf:params:jmap:core",
			"https://www.fastmail.com/dev/maskedemail",
		},
		MethodCalls: []MethodCall{r},
	}

	sendRes, err := client.sendRequest(&cmer)
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
	}(sendRes)
}
