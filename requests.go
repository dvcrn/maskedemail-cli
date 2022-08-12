package main

import "encoding/json"

type APIRequest struct {
	Using       []string     `json:"using,omitempty"`
	MethodCalls []MethodCall `json:"methodCalls,omitempty"`
}

type MethodCall struct {
	MethodName string
	Payload    interface{}
	Payload2   string
}

// MarshalJSON marshals a MethodCall into the format needed by the Fastmail API
// eg. ["MaskedEmail/set", payload, "0"].
func (r *MethodCall) MarshalJSON() ([]byte, error) {
	payloadJsonData, err := json.Marshal([]interface{}{r.MethodName, r.Payload, r.Payload2})
	if err != nil {
		return nil, err
	}

	return payloadJsonData, nil
}

/**
{
  "using": [
    "urn:ietf:params:jmap:core",
    "https://www.fastmail.com/dev/maskedemail"
  ],
  "methodCalls": [
    [
      "MaskedEmail/set",
      {
        "accountId": "xxxx",
        "create": {
          "onepassword": {
            "forDomain": "https://www.facebook.com"
          }
        }
      },
      "0"
    ]
  ]
}
*/

type CreatePayload struct {
	ForDomain string `json:"forDomain"`
	State     string `json:"state,omitempty"`
}

type MethodCallCreate struct {
	AccountID string                   `json:"accountId,omitempty"`
	Create    map[string]CreatePayload `json:"create,omitempty"`
}

type UpdateState struct {
	State string `json:"state,omitempty"`
}

// NewMethodCallCreate creates a new method call to create a new maskedemail.
// accID is the users account ID.
// appName is the name to identify the app that created the maskedemail.
// domain is the label to identify where the email is intended for.
func NewMethodCallCreate(accID, appName, domain string, state string) MethodCallCreate {
	mesp := MethodCallCreate{}
	mesp.AccountID = accID
	mesp.Create = map[string]CreatePayload{
		appName: {
			ForDomain: domain,
			State:     state,
		},
	}

	return mesp
}

type MethodCallUpdate struct {
	AccountID string                 `json:"accountId,omitempty"`
	Update    map[string]UpdateState `json:"update,omitempty"`
}

// NewMethodCallUpdateState creates a new method call to update a maskedemail.
// This is for example used when a temporary email is converted into a finalized one.
func NewMethodCallUpdateState(accID, alias string) MethodCallUpdate {
	mesp := MethodCallUpdate{}
	mesp.AccountID = accID
	mesp.Update = map[string]UpdateState{
		alias: {
			State: "enabled",
		},
	}

	return mesp
}
