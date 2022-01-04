package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/mitchellh/mapstructure"
)

const apiEndpoint = "https://api.fastmail.com/jmap/api/"

type Client struct {
	auth         string
	refreshToken string
	clientID     string
	accID        string
	appName      string
}

func NewClient(accID, token, appName, clientID string) *Client {
	return &Client{
		accID:        accID,
		auth:         "",
		refreshToken: token,
		appName:      appName,
		clientID:     clientID,
	}
}

func (client *Client) sendRequest(r *APIRequest) (*APIResponse, error) {
	reqJson, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(reqJson))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", fmt.Sprintf("Bearer %s", client.auth))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

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

// CreateMaskedEmail creates a new masked email for the given forDomain domain.
// If `enabled` is set to false, will only create a pending email and needs to be confirmed before it's usable.
func (client *Client) CreateMaskedEmail(forDomain string, enabled bool) (*MethodResponseCreateItem, error) {
	state := ""
	if enabled {
		state = "enabled"
	}

	mc := MethodCall{
		MethodName: "MaskedEmail/set",
		Payload:    NewMethodCallCreate(client.accID, client.appName, forDomain, state),
		Payload2:   "0",
	}

	request := APIRequest{
		Using: []string{
			"urn:ietf:params:jmap:core",
			"https://www.fastmail.com/dev/maskedemail",
		},
		MethodCalls: []MethodCall{mc},
	}

	res, err := client.sendRequest(&request)
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

func (client *Client) RefreshToken() (*RefreshTokenResponse, error) {
	val := url.Values{}
	val.Add("grant_type", "refresh_token")
	val.Add("client_id", client.clientID)
	val.Add("refresh_token", client.refreshToken)

	res, err := http.PostForm("https://api.fastmail.com/oauth/refresh", val)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(res.Body)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%s\n", buf.String())

	var tokenRes RefreshTokenResponse
	err = json.Unmarshal(buf.Bytes(), &tokenRes)
	if err != nil {
		return nil, err
	}

	client.auth = tokenRes.AccessToken

	return &tokenRes, nil
}

func (client *Client) ConfirmMaskedEmail(emailID string) (*MethodResponseCreateItem, error) {
	r := MethodCall{
		MethodName: "MaskedEmail/set",
		Payload:    NewMethodCallUpdateState(*flagAccountID, emailID),
		Payload2:   "0",
	}

	apiRequest := APIRequest{
		Using: []string{
			"urn:ietf:params:jmap:core",
			"https://www.fastmail.com/dev/maskedemail",
		},
		MethodCalls: []MethodCall{r},
	}

	res, err := client.sendRequest(&apiRequest)
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
