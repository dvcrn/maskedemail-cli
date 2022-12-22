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

const flagNameEmail string = "email"
const flagNameDomain string = "domain"
const flagNameDesc string = "desc"
const flagNameEnabled string = "enabled"

var flagAppname = flag.String("appname", os.Getenv(envAppVarName), "the appname to identify the creator (or "+envAppVarName+" env) (default: maskedemail-cli)")
var flagToken = flag.String("token", "", "the token to authenticate with (or "+envTokenVarName+" env)")
var flagAccountID = flag.String("accountid", os.Getenv(envAccountIdVarName), "fastmail account id (or "+envAccountIdVarName+" env)")


var listCmd = flag.NewFlagSet("list", flag.ExitOnError)
var flagShowDeleted = listCmd.Bool("show-deleted", false, "when enabled even deleted emails are shown, (default: false)")

var createCmd = flag.NewFlagSet("create", flag.ExitOnError)
var flagCreateDomain = createCmd.String(flagNameDomain, "", "domain for the masked email")
var flagCreateDescription = createCmd.String(flagNameDesc, "", "description for the masked email")
var flagCreateEnabled = createCmd.Bool(flagNameEnabled, true, "state of the masked email (default: true)")

var updateCmd = flag.NewFlagSet("update", flag.ExitOnError)
var flagUpdateEmail = updateCmd.String(flagNameEmail, "", "the masked email to update (required)")
var flagUpdateDomain = updateCmd.String(flagNameDomain, "", "domain for the masked email")
var flagUpdateDescription = updateCmd.String(flagNameDesc, "", "description for the masked email")

var action actionType = actionTypeUnknown
var envToken string

type actionType string

const (
	actionTypeUnknown            = ""
	actionTypeCreate             = "create"
	actionTypeSession            = "session"
	actionTypeDisable            = "disable"
	actionTypeEnable             = "enable"
	actionTypeDelete             = "delete"
	actionTypeUpdate             = "update"
	actionTypeList               = "list"
	defaultAppname               = "maskedemail-cli"
)

func isFlagPassed(set flag.FlagSet, name string) bool {
    found := false
    //fmt.Printf("name: %s\n", name)
    set.Visit(func(f *flag.Flag) {
    //	fmt.Printf("f.Name: %s\n", f.Name)
        if f.Name == name {
            found = true
        }
    })
    return found
}

func init() {
	flag.Parse()
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		fmt.Println("Flags:")
		flag.PrintDefaults()
		fmt.Println("")
		fmt.Println("Commands:")
		fmt.Println("  maskedemail-cli create [-domain \"<domain>\"] [-desc \"<description>\"] [-enabled=true|false (default: true)]")
		fmt.Println("  maskedemail-cli enable <maskedemail>")
		fmt.Println("  maskedemail-cli disable <maskedemail>")
		fmt.Println("  maskedemail-cli delete <maskedemail>")
		fmt.Println("  maskedemail-cli update -email <maskedemail> [-domain \"<domain>\"] [-desc \"<description>\"]")
		fmt.Println("  maskedemail-cli session")
		fmt.Println("  maskedemail-cli list [-show-deleted]")
	}

	if len(flag.Args()) < 1 {
		log.Println("no argument given. currently supported: create, session, enable, disable, delete, update, list")
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

	case "delete":
		action = actionTypeDelete

	case "list":
		action = actionTypeList

	case "update":
		action = actionTypeUpdate
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
		// parse command-specific args
		createCmd.Parse(os.Args[2:])

		domain := strings.TrimSpace(*flagCreateDomain)
		description := strings.TrimSpace(*flagCreateDescription)

		session, err := client.Session()
		if err != nil {
			log.Fatalf("initializing session: %v", err)
		}

		createRes, err := client.CreateMaskedEmail(session, *flagAccountID, domain, *flagCreateEnabled, description)
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

	case actionTypeDelete:
		if flag.Arg(1) == "" {
			log.Fatalln("Usage: delete <email>")
		}

		session, err := client.Session()
		if err != nil {
			log.Fatalf("initializing session: %v", err)
		}

		_, err = client.DeleteMaskedEmail(session, *flagAccountID, flag.Arg(1))
		if err != nil {
			log.Fatalf("err while deleting maskedemail: %v", err)
		}

		fmt.Printf("deleted maskedemail: %s\n", flag.Arg(1))

	case actionTypeList:
		// parse command-specific args
		listCmd.Parse(os.Args[2:])

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

			// HACK: trim space here is for hack to deal with possible empty strings
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", email.Email, strings.TrimSpace(email.Domain), strings.TrimSpace(email.Description), email.State, email.LastMessageAt, email.CreatedAt)
		}
		w.Flush()

	case actionTypeUpdate:
		// parse command-specific args
		updateCmd.Parse(os.Args[2:])

		maskedemail := strings.TrimSpace(*flagUpdateEmail)
		domain := strings.TrimSpace(*flagUpdateDomain)
		description := strings.TrimSpace(*flagUpdateDescription)

		// email arg is required
		if !isFlagPassed(*updateCmd, flagNameEmail) || (maskedemail == "") {
			updateCmd.Usage()
			os.Exit(1)
		}

		session, err := client.Session()
		if err != nil {
			log.Fatalf("initializing session: %v", err)
		}

		fields := pkg.NewUpdateFields(isFlagPassed(*updateCmd, flagNameDomain),
									  domain,
									  isFlagPassed(*updateCmd, flagNameDesc),
									  description)

		_, err = client.UpdateInfo(session, *flagAccountID, maskedemail, fields)
		if err != nil {
			log.Fatalf("err updating maskedemail: %v", err)
		}

		fmt.Printf("updated %s\n", maskedemail)

	default:
		fmt.Println("action not found")
		flag.Usage()
		os.Exit(1)
	}
}
