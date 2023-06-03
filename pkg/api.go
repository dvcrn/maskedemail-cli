package pkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/mitchellh/mapstructure"
)

const (
	// sessionEndpoint is used to auto-discover the main API endpoint
	sessionEndpoint = "https://api.fastmail.com/jmap/session"

	// MaskedEmailCapabilityURI is the capability URI for the Masked Email
	// feature within the JMAP API.
	//
	// https://beta.fastmail.com/developer/maskedemail/
	MaskedEmailCapabilityURI = "https://www.fastmail.com/dev/maskedemail"
)

// errNoAccountID is returned if an account ID is not explicitly provided and
// a primary account is not found for the required capability URI.
var errNoAccountID = errors.New("no account specified and no default account for masked email")

// Session contains server metadata information as well as the available
// accounts for the provided credentials.
type Session interface {
	// ApiEndpoint is the URL to use for JMAP API requests.
	ApiEndpoint() string

	// AccountHasCapability returns true if the specified account ID has access to
	// the specified capability URI.
	AccountHasCapability(accID string, capabilityURI string) bool

	// DefaultAccountForCapability returns the default account ID (if any) for
	// the given capability URI.
	DefaultAccountForCapability(capabilityURI string) string
}

type Client struct {
	auth     string
	clientID string
	appName  string
}

func NewClient(token, appName, clientID string) *Client {
	return &Client{
		auth:     token,
		appName:  appName,
		clientID: clientID,
	}
}

// doRequest adds common headers and executes the HTTP request.
func (client *Client) doRequest(req *http.Request) (*http.Response, error) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", fmt.Sprintf("Bearer %s", client.auth))
	return http.DefaultClient.Do(req)
}

func (client *Client) sendRequest(session Session, r *APIRequest) (*APIResponse, error) {
	reqJson, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", session.ApiEndpoint(), bytes.NewReader(reqJson))
	if err != nil {
		return nil, err
	}

	res, err := client.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(res.Body)
	if err != nil {
		return nil, err
	}

	var apiRes APIResponse
	err = json.Unmarshal(buf.Bytes(), &apiRes)
	if err != nil {
		return nil, err
	}

	return &apiRes, nil
}

// Session queries the JMAP auto-discovery endpoint for details about the
// server and available accounts.
func (client *Client) Session() (*SessionResource, error) {
	req, err := http.NewRequest(http.MethodGet, sessionEndpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	jsonBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var session SessionResource
	if err := json.Unmarshal(jsonBody, &session); err != nil {
		return nil, err
	}

	return &session, nil
}

func (client *Client) accIDOrDefault(session Session, accID string) (string, error) {
	if accID != "" {
		return accID, nil
	}

	accID = session.DefaultAccountForCapability(MaskedEmailCapabilityURI)
	if accID == "" {
		return "", errNoAccountID
	}

	return accID, nil
}

// CreateMaskedEmail creates a new masked email for the given domain.
//
// If `accID` is the empty string, the primary account for Masked Email will be
// used.
//
// If `enabled` is set to false, will only create a pending email and needs to be confirmed before it's usable.
func (client *Client) CreateMaskedEmail(
	session Session,
	accID string,
	domain string,
	description string,
	emailPrefix string,
	enabled bool,
) (*MaskedEmail, error) {
	state := ""
	if enabled {
		state = "enabled"
	}

	accID, err := client.accIDOrDefault(session, accID)
	if err != nil {
		return nil, err
	}

	mc := MethodCall{
		MethodName: "MaskedEmail/set",
		Payload:    NewMethodCallCreate(accID, client.appName, domain, state, description, emailPrefix),
		Payload2:   "0",
	}

	request := APIRequest{
		Using: []string{
			"urn:ietf:params:jmap:core",
			MaskedEmailCapabilityURI,
		},
		MethodCalls: []MethodCall{mc},
	}

	res, err := client.sendRequest(session, &request)
	if err != nil {
		return nil, err
	}

	var pl MethodResponseMaskedEmailSet
	err = mapstructure.Decode(res.MethodResponsesParsed[0].Payload, &pl)
	if err != nil {
		return nil, err
	}

	created, err := pl.GetCreatedItem()
	if err != nil {
		return nil, err
	}

	return &created, nil
}

func (client *Client) UpdateMaskedEmail(
	session Session,
	accID string,
	emailID string,
	updateOpts ...UpdateOption,
) (*MethodResponseMaskedEmailSet, error) {

	accID, err := client.accIDOrDefault(session, accID)
	if err != nil {
		return nil, err
	}

	// handle more than one with a value at once?
	var payload MethodCallUpdate = NewMethodCallUpdate(accID, emailID, updateOpts...)

	r := MethodCall{
		MethodName: "MaskedEmail/set",
		Payload:    payload,
		Payload2:   "0",
	}

	apiRequest := APIRequest{
		Using: []string{
			"urn:ietf:params:jmap:core",
			MaskedEmailCapabilityURI,
		},
		MethodCalls: []MethodCall{r},
	}

	res, err := client.sendRequest(session, &apiRequest)
	if err != nil {
		return nil, err
	}

	var pl MethodResponseMaskedEmailSet
	err = mapstructure.Decode(res.MethodResponsesParsed[0].Payload, &pl)
	if err != nil {
		return nil, err
	}

	// TODO: fix return value
	pl.GetCreatedItem()

	return nil, nil
}

func (client *Client) LookupMaskedEmailID(
	session Session,
	accID string,
	email string,
) (string, error) {
	allAliases, err := client.GetAllMaskedEmails(session, accID, true)
	if err != nil {
		return "", err
	}

	// find the alias to disable
	var alias *MaskedEmail
	for _, a := range allAliases {
		if a.Email == email {
			alias = a
			break
		}
	}

	if alias == nil {
		return "", errors.New(fmt.Sprintf("maskedemail %s not found", email))
	}

	return alias.ID, nil
}

func (client *Client) EnableMaskedEmail(
	session Session,
	accID string,
	email string,
) (*MethodResponseMaskedEmailSet, error) {
	emailID, err := client.LookupMaskedEmailID(session, accID, email)

	if err != nil {
		return nil, err
	}

	return client.UpdateMaskedEmail(session, accID, emailID, WithUpdateState(MaskedEmailStateEnabled))
}

func (client *Client) DisableMaskedEmail(
	session Session,
	accID string,
	email string,
) (*MethodResponseMaskedEmailSet, error) {

	emailID, err := client.LookupMaskedEmailID(session, accID, email)

	if err != nil {
		return nil, err
	}

	return client.UpdateMaskedEmail(session, accID, emailID, WithUpdateState(MaskedEmailStateDisabled))
}

func (client *Client) DeleteMaskedEmail(
	session Session,
	accID string,
	email string,
) (*MethodResponseMaskedEmailSet, error) {

	emailID, err := client.LookupMaskedEmailID(session, accID, email)

	if err != nil {
		return nil, err
	}

	return client.UpdateMaskedEmail(session, accID, emailID, WithUpdateState(MaskedEmailStateDeleted))
}

func (client *Client) UpdateInfo(
	session Session,
	accID string,
	email string,
	updateOpts ...UpdateOption,
) (*MethodResponseMaskedEmailSet, error) {

	emailID, err := client.LookupMaskedEmailID(session, accID, email)

	if err != nil {
		return nil, err
	}

	return client.UpdateMaskedEmail(session, accID, emailID, updateOpts...)
}

func (client *Client) GetAllMaskedEmails(
	session Session,
	accID string,
	includeDeleted bool,
) ([]*MaskedEmail, error) {
	accID, err := client.accIDOrDefault(session, accID)
	if err != nil {
		return nil, err
	}

	r := MethodCall{
		MethodName: "MaskedEmail/get",
		Payload:    NewMethodCallGetAll(accID),
		Payload2:   "0",
	}

	apiRequest := APIRequest{
		Using: []string{
			"urn:ietf:params:jmap:core",
			MaskedEmailCapabilityURI,
		},
		MethodCalls: []MethodCall{r},
	}

	res, err := client.sendRequest(session, &apiRequest)
	if err != nil {
		return nil, err
	}

	var pl MethodResponseGetAll
	err = mapstructure.Decode(res.MethodResponsesParsed[0].Payload, &pl)
	if err != nil {
		return nil, err
	}

	out := []*MaskedEmail{}
	for _, item := range pl.List {
		// skip deleted masked emails unless flag to show is passed
		if item.State == "deleted" && !includeDeleted {
			continue
		}

		out = append(out, item)
	}

	return out, nil
}
