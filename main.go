package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var flagDomain = flag.String("domain", "", "the domain to create the alias for")
var flagAppname = flag.String("appname", "maskedemail-cli", "the appname to identify the creator")
var flagToken = flag.String("token", "", "the token to authenticate with")
var flagAccountID = flag.String("accountid", "", "fastmail account id")

func init() {
	flag.Parse()

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
	client := NewClient(*flagAccountID, *flagToken, *flagAppname)

	createRes, err := client.CreateMaskedEmail(*flagDomain)
	if err != nil {
		log.Fatalf("err while creating maskedemail: %v", err)
	}

	// _, err = client.ConfirmMaskedEmail(createRes.ID)
	// if err != nil {
	// 	log.Fatalf("err while confirming maskedemail: %v", err)
	// }

	fmt.Println(createRes.Email)
}
