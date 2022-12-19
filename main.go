package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/dvcrn/maskedemail-cli/pkg"
)

const envTokenVarName string = "MASKEDEMAIL_TOKEN"
const envAppVarName string = "MASKEDEMAIL_APPNAME"
const envAccountIdVarName string = "MASKEDEMAIL_ACCOUNTID"

var flagAppname = flag.String("appname", os.Getenv(envAppVarName), "the appname to identify the creator (or "+envAppVarName+" env) (default: maskedemail-cli)")
var flagToken = flag.String("token", "", "the token to authenticate with (or "+envTokenVarName+" env)")
var flagAccountID = flag.String("accountid", os.Getenv(envAccountIdVarName), "fastmail account id (or "+envAccountIdVarName+" env)")
var flagShowDeleted = flag.Bool("show-deleted", false, "when enabled even deleted emails are shown, (default: false)")
var action actionType = actionTypeUnknown
var envToken string

type actionType string

const (
	actionTypeUnknown = ""
	actionTypeCreate  = "create"
	actionTypeSession = "session"
	actionTypeDisable = "disable"
	actionTypeEnable  = "enable"
	actionTypeList    = "list"
	defaultAppname    = "maskedemail-cli"
)

func init() {
	flag.Parse()
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Println("Flags:")
		flag.PrintDefaults()
		fmt.Println("")
		fmt.Println("Commands:")
		fmt.Println("  maskedemail-cli create <domain> <description>")
		fmt.Println("  maskedemail-cli enable <maskedemail>")
		fmt.Println("  maskedemail-cli disable <maskedemail>")
		fmt.Println("  maskedemail-cli session")
		fmt.Println("  maskedemail-cli list")
	}

	if len(flag.Args()) < 1 {
		log.Println("no argument given. currently supported: create, session, disable, enable")
		flag.Usage()
		os.Exit(1)
	}

	// CLI parameter have precedence over ENV variables
	if *flagToken == "" {
		envToken = os.Getenv(envTokenVarName)
		if envToken != "" {
			*flagToken = envToken
		} else {
			flag.Usage()
			os.Exit(1)
		}
	}

	if *flagAppname == "" {
		*flagAppname = defaultAppname
	}

	switch strings.ToLower(flag.Arg(0)) {
	case
		"create":
		action = actionTypeCreate

	case "session":
		action = actionTypeSession

	case "disable":
		action = actionTypeDisable

	case "enable":
		action = actionTypeEnable

	case "list":
		action = actionTypeList
	}
}

func main() {
	client := pkg.NewClient(*flagToken, *flagAppname, "35c941ae")

	switch action {
	case actionTypeSession:
		session, err := client.Session()
		if err != nil {
			log.Fatalf("fetching session: %v", err)
		}
		var accIDs []string
		for accID := range session.Accounts {
			if *flagAccountID != "" && *flagAccountID != accID {
				continue
			}
			accIDs = append(accIDs, accID)
		}

		primaryAccountID := session.PrimaryAccounts[pkg.MaskedEmailCapabilityURI]
		sort.Slice(
			accIDs,
			func(i, j int) bool {
				if primaryAccountID == accIDs[i] {
					return true
				}
				return accIDs[i] < accIDs[j]
			},
		)
		for _, accID := range accIDs {
			isPrimary := primaryAccountID == accID
			isEnabled := session.AccountHasCapability(accID, pkg.MaskedEmailCapabilityURI)

			fmt.Printf(
				"%s [%s] (primary: %t, enabled: %t)\n",
				session.Accounts[accID].Name,
				accID,
				isPrimary,
				isEnabled,
			)
		}

	case actionTypeCreate:

		if (len(flag.Args()) != 3) {
			log.Fatalln("Usage: create <domain> <description>")
		}

		if (strings.TrimSpace(flag.Arg(1)) == "") && (strings.TrimSpace(flag.Arg(2)) == "") {
			fmt.Println("Usage: create <domain> <description>")
			log.Fatalln("At least one of <domain> and <description> are required")
		}

		session, err := client.Session()
		if err != nil {
			log.Fatalf("initializing session: %v", err)
		}

		createRes, err := client.CreateMaskedEmail(session, *flagAccountID, flag.Arg(1), true, flag.Arg(2))
		if err != nil {
			log.Fatalf("err while creating maskedemail: %v", err)
		}

		fmt.Println(createRes.Email)

	case actionTypeDisable:
		if flag.Arg(1) == "" {
			log.Fatalln("Usage: disable <maskedemail>")
		}

		session, err := client.Session()
		if err != nil {
			log.Fatalf("initializing session: %v", err)
		}

		_, err = client.DisableMaskedEmail(session, *flagAccountID, flag.Arg(1))
		if err != nil {
			log.Fatalf("err disabling maskedemail: %v", err)
		}

		fmt.Printf("disabled email: %s\n", flag.Arg(1))

	case actionTypeEnable:
		if flag.Arg(1) == "" {
			log.Fatalln("Usage: enable <email>")
		}

		session, err := client.Session()
		if err != nil {
			log.Fatalf("initializing session: %v", err)
		}

		_, err = client.EnableMaskedEmail(session, *flagAccountID, flag.Arg(1))
		if err != nil {
			log.Fatalf("err while updating maskedemail: %v", err)
		}

		fmt.Printf("enabled maskedemail: %s\n", flag.Arg(1))

	case actionTypeList:
		session, err := client.Session()
		if err != nil {
			log.Fatalf("initializing session: %v", err)
		}

		maskedEmails, err := client.GetAllMaskedEmails(session, *flagAccountID)
		if err != nil {
			log.Fatalf("err while creating maskedemail: %v", err)
		}

		w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
		fmt.Fprintln(w, "Masked Email\tFor Domain\tDescription\tState\tLast Email At\tCreated At")
		for _, email := range maskedEmails {
			if email.State == "deleted" && !*flagShowDeleted {
				continue
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", email.Email, email.ForDomain, email.Description, email.State, email.LastMessageAt, email.CreatedAt)
		}
		w.Flush()

	default:
		fmt.Println("action not found")
		flag.Usage()
		os.Exit(1)
	}
}
