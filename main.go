package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var flagAppname = flag.String("appname", "maskedemail-cli", "the appname to identify the creator")
var flagToken = flag.String("token", "", "the token to authenticate with")
var flagAccountID = flag.String("accountid", "", "fastmail account id")
var flagUseRefresh = flag.Bool("refresh", false, "whether the token is a refresh token")
var action actionType = actionTypeUnknown

type actionType string

const (
	actionTypeUnknown = ""
	actionTypeCreate  = "create"
	actionTypeAuth    = "auth"
)

func init() {
	flag.Parse()
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Println("Flags:")
		flag.PrintDefaults()
		fmt.Println("")
		fmt.Println("Commands:")
		fmt.Println("  maskedemail-cli create <domain>")
		fmt.Println("  maskedemail-cli auth <email> <password>")
	}

	if len(flag.Args()) < 1 {
		log.Println("no argument given. currently supported: create, auth")
		flag.Usage()
		os.Exit(0)
	}

	switch strings.ToLower(flag.Arg(0)) {
	case
		"create":
		action = actionTypeCreate

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

	case "auth":
		action = actionTypeAuth
	}
}

func main() {
	client := NewClient(*flagAccountID, *flagToken, *flagAppname, "35c941ae")

	switch action {
	case actionTypeAuth:
		if len(flag.Args()) != 3 {
			log.Println("Usage: auth <email> <password>")
			return
		}

		res, err := Authenticate(flag.Args()[1], flag.Args()[2])
		if err != nil {
			fmt.Println("authentication failed:", err)
			return
		}

		func(v interface{}) {
			j, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				fmt.Printf("%v\n", err)
				return
			}
			buf := bytes.NewBuffer(j)
			fmt.Printf("%v\n", buf.String())
		}(res)

	case actionTypeCreate:
		if flag.Arg(1) == "" {
			log.Println("Usage: create <domain>")
			return
		}

		if *flagUseRefresh {
			_, err := client.RefreshToken()
			if err != nil {
				panic(err)
			}
		}

		createRes, err := client.CreateMaskedEmail(flag.Arg(1), true)
		if err != nil {
			log.Fatalf("err while creating maskedemail: %v", err)
		}

		fmt.Print(createRes.Email)

	default:
		fmt.Println("action not found")
		flag.Usage()
	}
}
