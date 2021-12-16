package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"

	"github.com/mitchellh/mapstructure"
)

type MaskedEmailSetResPayloadCreatedRow struct {
	CreatedAt     string `mapstructure:"createdAt"`
	CreatedBy     string `mapstructure:"createdBy"`
	Description   string `mapstructure:"description"`
	Email         string `mapstructure:"email"`
	ID            string `mapstructure:"id"`
	LastMessageAt string `mapstructure:"lastMessageAt"`
	State         string `mapstructure:"state"`
	URL           string `mapstructure:"url"`
}

type MaskedEmailSetResPayload struct {
	AccountID string                                        `mapstructure:"accountId"`
	Created   map[string]MaskedEmailSetResPayloadCreatedRow `mapstructure:"created"`
}

type Res struct {
	MethodName string
	Payload    interface{}
	Payload2   string
}

type GenericResponse struct {
	LatestClientVersion string          `json:"latestClientVersion,omitempty"`
	MethodResponses     [][]interface{} `json:"methodResponses,omitempty"`
	ParsedResponses     []Res
	SessionState        string `json:"sessionState,omitempty"`
}

func (gr *GenericResponse) UnmarshalJSON(b []byte) error {
	type person2 GenericResponse
	if err := json.Unmarshal(b, (*person2)(gr)); err != nil {
		return err
	}

	responses := []Res{}
	for _, res := range gr.MethodResponses {
		r := Res{}
		r.MethodName = res[0].(string)
		r.Payload = res[1]
		r.Payload2 = res[2].(string)

		responses = append(responses, r)
	}

	gr.ParsedResponses = responses
	return nil
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

type Req struct {
	MethodName string
	Payload    interface{}
	Payload2   string
}

func (r *Req) ToJSON() ([]byte, error) {
	// eg. ["MaskedEmail/set", payload, 0]

	payloadJsonData, err := json.Marshal([]interface{}{r.MethodName, r.Payload, r.Payload2})
	if err != nil {
		return nil, err
	}

	return payloadJsonData, nil
}

func (r *Req) MarshalJSON() ([]byte, error) {
	return r.ToJSON()
}

type ForDomain struct {
	ForDomain string `json:"forDomain"`
}
type MaskedEmailSetParams struct {
	AccountID string               `json:"accountId,omitempty"`
	Create    map[string]ForDomain `json:"create,omitempty"`
}

type State struct {
	State string `json:"state,omitempty"`
}

func NewMaskedEmailSetParams(accID, domain string) MaskedEmailSetParams {
	mesp := MaskedEmailSetParams{}
	mesp.AccountID = accID
	mesp.Create = map[string]ForDomain{
		"myapp": {
			ForDomain: domain,
		},
	}

	return mesp
}

type MaskedEmailUpdateParams struct {
	AccountID string           `json:"accountId,omitempty"`
	Update    map[string]State `json:"update,omitempty"`
}

func NewMaskedEmailUpdateParams(accID, alias string) MaskedEmailUpdateParams {
	mesp := MaskedEmailUpdateParams{}
	mesp.AccountID = accID
	mesp.Update = map[string]State{
		alias: {
			State: "enabled",
		},
	}

	return mesp
}

// ["MaskedEmail/set", idkyet, 0]

const auth = ""
const accID = ""
const apiHost = "api.fastmail.com"
const apiEndpoint = "https://api.fastmail.com/jmap/api/"

type CreatedMaskedEmailRequest struct {
	Using       []string `json:"using,omitempty"`
	MethodCalls []Req    `json:"methodCalls,omitempty"`
}

func main() {
	flag.Parse()

	r := Req{
		MethodName: "MaskedEmail/set",
		Payload:    NewMaskedEmailSetParams(accID, "hoge.com"),
		Payload2:   "0",
	}

	cmer := CreatedMaskedEmailRequest{
		Using: []string{
			"urn:ietf:params:jmap:core",
			"https://www.fastmail.com/dev/maskedemail",
		},
		MethodCalls: []Req{r},
	}

	reqJson, err := json.Marshal(cmer)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%v", string(reqJson))

	req, err := http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(reqJson))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", fmt.Sprintf("Bearer %s", auth))

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	buf := &bytes.Buffer{}
	buf.ReadFrom(res.Body)

	var genRe GenericResponse
	json.Unmarshal(buf.Bytes(), &genRe)

	// TODO: for debug. remove me.
	func(v interface{}) {
		j, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}
		buf := bytes.NewBuffer(j)
		fmt.Printf("%v\n", buf.String())
	}(genRe.ParsedResponses[0].Payload)

	var pl MaskedEmailSetResPayload
	err = mapstructure.Decode(genRe.ParsedResponses[0].Payload, &pl)
	if err != nil {
		panic(err)
	}

	// TODO: for debug. remove me.
	func(v interface{}) {
		j, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}
		buf := bytes.NewBuffer(j)
		fmt.Printf("%v\n", buf.String())
	}(pl)

	r = Req{
		MethodName: "MaskedEmail/set",
		Payload:    NewMaskedEmailUpdateParams(accID, pl.Created["myapp"].ID),
		Payload2:   "0",
	}

	cmer = CreatedMaskedEmailRequest{
		Using: []string{
			"urn:ietf:params:jmap:core",
			"https://www.fastmail.com/dev/maskedemail",
		},
		MethodCalls: []Req{r},
	}

	fmt.Println("second res")
	// TODO: for debug. remove me.
	func(v interface{}) {
		j, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}
		buf := bytes.NewBuffer(j)
		fmt.Printf("%v\n", buf.String())
	}(cmer)

	reqJson, err = json.Marshal(cmer)
	if err != nil {
		panic(err)
	}

	req, err = http.NewRequest("POST", apiEndpoint, bytes.NewBuffer(reqJson))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authorization", fmt.Sprintf("Bearer %s", auth))

	res, err = http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	// TODO: for debug. remove me.
	buf = &bytes.Buffer{}
	buf.ReadFrom(res.Body)

	fmt.Printf("%v", buf.String())
}
