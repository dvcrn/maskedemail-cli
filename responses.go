package main

import (
	"encoding/json"
	"errors"
)

type MethodResponse struct {
	MethodName string
	Payload    interface{}
	Payload2   string
}

type APIResponse struct {
	LatestClientVersion   string          `json:"latestClientVersion,omitempty"`
	MethodResponses       [][]interface{} `json:"methodResponses,omitempty"`
	MethodResponsesParsed []MethodResponse
	SessionState          string `json:"sessionState,omitempty"`
}

func (gr *APIResponse) UnmarshalJSON(b []byte) error {
	type apiResponse2 APIResponse
	if err := json.Unmarshal(b, (*apiResponse2)(gr)); err != nil {
		return err
	}

	responses := []MethodResponse{}
	for _, res := range gr.MethodResponses {
		r := MethodResponse{}
		r.MethodName = res[0].(string)
		r.Payload = res[1]
		r.Payload2 = res[2].(string)

		responses = append(responses, r)
	}

	gr.MethodResponsesParsed = responses
	return nil
}

type MaskedEmail struct {
	CreatedAt     string `mapstructure:"createdAt"`
	CreatedBy     string `mapstructure:"createdBy"`
	Description   string `mapstructure:"description"`
	Email         string `mapstructure:"email"`
	ID            string `mapstructure:"id"`
	LastMessageAt string `mapstructure:"lastMessageAt"`
	State         string `mapstructure:"state"`
	URL           string `mapstructure:"url"`
	ForDomain     string `mapstructure:"forDomain"`
}

type MethodResponseMaskedEmailSet struct {
	AccountID string                 `mapstructure:"accountId"`
	Created   map[string]MaskedEmail `mapstructure:"created"`
	Updated   map[string]interface{} `mapstructure:"updated"`
	Destroyed []interface{}          `mapstructure:"destroyed"`
	NewState  interface{}            `mapstructure:"newState"`
	OldState  interface{}            `mapstructure:"oldState"`
}

func (cr *MethodResponseMaskedEmailSet) GetCreatedItem() (MaskedEmail, error) {
	for _, item := range cr.Created {
		return item, nil
	}

	return MaskedEmail{}, errors.New("no items returned")
}

type MethodResponseGetAll struct {
	AccountID string         `mapstructure:"accountId"`
	NotFound  []interface{}  `mapstructure:"notFound"`
	State     string         `mapstructure:"state"`
	List      []*MaskedEmail `mapstructure:"list"`
}

// Account is a collection of data in the JMAP API.
//
// https://jmap.io/spec-core.html#terminology
type Account struct {
	// Name is a user-friendly string to show when presenting content from this
	// account, e.g., the email address representing the owner of the account.
	Name string `json:"name"`
	// Capabilities is the set of capability URIs for the methods supported in
	// this account.
	Capabilities map[string]json.RawMessage `json:"accountCapabilities"`
}

// SessionResource gives details about the data and capabilities the server can
// provide to the client given those credentials.
//
// It is the initial response from a JMAP auto-discovery endpoint.
//
// https://jmap.io/spec-core.html#the-jmap-session-resource
type SessionResource struct {
	// Capabilities is an object specifying the capabilities of this server.
	// Each key is a URI for a capability supported by the server.
	Capabilities map[string]json.RawMessage `json:"capabilities"`
	// Accounts is a map of an account id to an Account object for each account
	// the user has access to.
	Accounts map[string]Account `json:"accounts"`
	// PrimaryAccounts is map of capability URIs (as found in
	// `accountCapabilities`) to the account id that is considered to be the
	// user's main or default account for data pertaining to that capability.
	PrimaryAccounts map[string]string `json:"primaryAccounts"`
	// ApiUrl is the URL to use for JMAP API requests.
	ApiUrl string `json:"apiUrl"`
}

var _ Session = &SessionResource{}

func (s *SessionResource) ApiEndpoint() string {
	return s.ApiUrl
}

func (s *SessionResource) DefaultAccountForCapability(capabilityURI string) string {
	return s.PrimaryAccounts[capabilityURI]
}

func (s *SessionResource) AccountHasCapability(accID string, capabilityURI string) bool {
	_, ok := s.Accounts[accID].Capabilities[capabilityURI]
	return ok
}
