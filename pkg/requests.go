package pkg

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
	Domain      string `json:"forDomain"`
	State       string `json:"state,omitempty"`
	Description string `json:"description"`
}

type MethodCallCreate struct {
	AccountID string                   `json:"accountId,omitempty"`
	Create    map[string]CreatePayload `json:"create,omitempty"`
}

type UpdatePayload struct {
	State string `json:"state,omitempty"`
	Domain string `json:"forDomain,omitempty"`
	Description string `json:"description,omitempty"`
}

// NewMethodCallCreate creates a new method call to create a new maskedemail.
// accID is the users account ID.
// appName is the name to identify the app that created the maskedemail.
// domain is the label to identify where the email is intended for.
// description is a description of the masked email
func NewMethodCallCreate(accID, appName, domain string, state string, description string) MethodCallCreate {
	mesp := MethodCallCreate{}
	mesp.AccountID = accID
	mesp.Create = map[string]CreatePayload{
		appName: {
			Domain:      domain,
			State:       state,
			Description: description,
		},
	}

	return mesp
}

type MaskedEmailState string

const (
	MaskedEmailStateEnabled  MaskedEmailState = "enabled"
	MaskedEmailStateDisabled                  = "disabled"
	MaskedEmailStateDeleted                   = "deleted"
)

type MethodCallUpdate struct {
	AccountID string                 `json:"accountId,omitempty"`
	Update    map[string]UpdatePayload `json:"update,omitempty"`
}

// NewMethodCallUpdateState creates a new method call to update a maskedemail.
// This is for example used when a temporary email is converted into a finalized one.
func NewMethodCallUpdateState(accID, alias string, state MaskedEmailState) MethodCallUpdate {
	mesp := MethodCallUpdate{}
	mesp.AccountID = accID
	mesp.Update = map[string]UpdatePayload{
		alias: {
			State: string(state),
		},
	}

	return mesp
}

// NewMethodCallUpdateDomain creates a new method call to update a maskedemail.
func NewMethodCallUpdateDomain(accID, alias string, domain string) MethodCallUpdate {
	mesp := MethodCallUpdate{}
	mesp.AccountID = accID

	// HACK: to make it appear the value is cleared out
	//  if left as empty string, the conversion to json's
	//  omitempty will not pass value in request and it won't get changed
	if domain == "" {
		domain = " "
	}

	mesp.Update = map[string]UpdatePayload{
		alias: {
			Domain: string(domain),
		},
	}

	return mesp
}

// NewMethodCallUpdateDescription creates a new method call to update a maskedemail.
func NewMethodCallUpdateDescription(accID, alias string, description string) MethodCallUpdate {
	mesp := MethodCallUpdate{}
	mesp.AccountID = accID

	// HACK: to make it appear the value is cleared out
	//  if left as empty string, the conversion to json's
	//  omitempty will not pass value in request and it won't get changed
	if description == "" {
		description = " "
	}

	mesp.Update = map[string]UpdatePayload{
		alias: {
			Description: string(description),
		},
	}

	return mesp
}

// MethodCallGetAll is a method call to get all maskedemails for a user.
// Request:
//    "methodCalls" : [
//      [
//         "MaskedEmail/get",
//         {
//            "accountId" : "xxx",
//            "ids" : null
//         },
//         "0"
//      ]
//   ],
//
// Response:
//   "methodResponses" : [
//      [
//         "MaskedEmail/get",
//         {
//            "accountId" : xxx",
//            "list" : [
//               {
//                  "createdAt" : "2021-09-29T23:02:05Z",
//                  "createdBy" : "",
//                  "description" : "Masked Email Example (yellow.asdfkjasdf)",
//                  "email" : "foo@bar.com",
//                  "forDomain" : "fastmail.com",
//                  "id" : "someid",
//                  "lastMessageAt" : "2021-09-29T23:02:06Z",
//                  "state" : "deleted",
//                  "url" : null
//               }, ...
//            ]
//         },
//      ]
//   ]
//
type MethodCallGetAll struct {
	AccountID string `json:"accountId,omitempty"`
}

func NewMethodCallGetAll(accID string) MethodCallGetAll {
	mesp := MethodCallGetAll{}
	mesp.AccountID = accID

	return mesp
}
