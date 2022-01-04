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

type MethodResponseCreateItem struct {
	CreatedAt     string `mapstructure:"createdAt"`
	CreatedBy     string `mapstructure:"createdBy"`
	Description   string `mapstructure:"description"`
	Email         string `mapstructure:"email"`
	ID            string `mapstructure:"id"`
	LastMessageAt string `mapstructure:"lastMessageAt"`
	State         string `mapstructure:"state"`
	URL           string `mapstructure:"url"`
}

type MethodResponseMaskedEmailSet struct {
	AccountID string                              `mapstructure:"accountId"`
	Created   map[string]MethodResponseCreateItem `mapstructure:"created"`
	Updated   map[string]interface{}              `mapstructure:"updated"`
	Destroyed []interface{}                       `mapstructure:"destroyed"`
	NewState  interface{}                         `mapstructure:"newState"`
	OldState  interface{}                         `mapstructure:"oldState"`
}

func (cr *MethodResponseMaskedEmailSet) GetCreatedItem() (MethodResponseCreateItem, error) {
	for _, item := range cr.Created {
		return item, nil
	}

	return MethodResponseCreateItem{}, errors.New("no items returned")
}

type RefreshTokenResponse struct {
	Scope        string `json:"scope,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	TokenType    string `json:"token_type,omitempty"`
	ExpiresIn    int    `json:"expires_in,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
}
