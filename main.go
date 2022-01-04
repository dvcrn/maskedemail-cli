package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var flagAppname = flag.String("appname", "maskedemail-cli", "the appname to identify the creator")
var flagToken = flag.String("token", "", "the refresh token to authenticate with")
var flagAccountID = flag.String("accountid", "", "fastmail account id")
var action actionType = actionTypeUnknown

type actionType string

const (
	actionTypeUnknown = ""
	actionTypeCreate  = "create"
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
	}

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

	if len(flag.Args()) < 1 {
		log.Println("no argument given. currently supported: create")
		flag.Usage()
		os.Exit(0)
	}

	switch strings.ToLower(flag.Arg(0)) {
	case
		"create":
		action = actionTypeCreate
	}
}

func main() {
	client := NewClient(*flagAccountID, *flagToken, *flagAppname, "35c941ae")

	switch action {
	case actionTypeCreate:
		if flag.Arg(1) == "" {
			log.Println("Usage: create <domain>")
			return
		}

		_, err := client.RefreshToken()
		if err != nil {
			panic(err)
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

	// _, err = client.ConfirmMaskedEmail(createRes.ID)
	// if err != nil {
	// 	log.Fatalf("err while confirming maskedemail: %v", err)
	// }

}
