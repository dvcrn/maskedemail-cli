package main

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

	// maskedEmailCapabilityURI is the capability URI for the Masked Email
	// feature within the JMAP API.
	//
	// https://beta.fastmail.com/developer/maskedemail/
	maskedEmailCapabilityURI = "https://www.fastmail.com/dev/maskedemail"
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

// CreateMaskedEmail creates a new masked email for the given forDomain domain.
//
// If `accID` is the empty string, the primary account for Masked Email will be
// used.
//
// If `enabled` is set to false, will only create a pending email and needs to be confirmed before it's usable.
func (client *Client) CreateMaskedEmail(
	session Session,
	accID string,
	forDomain string,
	enabled bool,
) (*MethodResponseCreateItem, error) {
	state := ""
	if enabled {
		state = "enabled"
	}

	if accID == "" {
		accID = session.DefaultAccountForCapability(maskedEmailCapabilityURI)
		if accID == "" {
			return nil, errNoAccountID
		}
	}

	mc := MethodCall{
		MethodName: "MaskedEmail/set",
		Payload:    NewMethodCallCreate(accID, client.appName, forDomain, state),
		Payload2:   "0",
	}

	request := APIRequest{
		Using: []string{
			"urn:ietf:params:jmap:core",
			maskedEmailCapabilityURI,
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

func (client *Client) ConfirmMaskedEmail(
	session Session,
	accID string,
	emailID string,
) (*MethodResponseCreateItem, error) {
	if accID == "" {
		accID = session.DefaultAccountForCapability(maskedEmailCapabilityURI)
		if accID == "" {
			return nil, errNoAccountID
		}
	}

	r := MethodCall{
		MethodName: "MaskedEmail/set",
		Payload:    NewMethodCallUpdateState(accID, emailID),
		Payload2:   "0",
	}

	apiRequest := APIRequest{
		Using: []string{
			"urn:ietf:params:jmap:core",
			maskedEmailCapabilityURI,
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

	return nil, nil
}
