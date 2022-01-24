package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
)

const authenticateURL = "https://www.fastmail.com/jmap/authenticate/"

type AuthenticateUsernameRequest struct {
	Username string `json:"username,omitempty"`
}

type AuthenticateUsernameResponse struct {
	MayTrustDevice bool `json:"mayTrustDevice,omitempty"`
	Methods        []struct {
		Type string `json:"type,omitempty"`
	} `json:"methods,omitempty"`
	LoginId string `json:"loginId,omitempty"`
}

type AuthenticatePasswordRequest struct {
	LoginId  string `json:"loginId,omitempty"`
	Remember bool   `json:"remember"`
	Type     string `json:"type,omitempty"`
	Value    string `json:"value,omitempty"`
}

type AuthenticatePasswordResponse struct {
	LoginId        string `json:"loginId,omitempty"`
	MayTrustDevice bool   `json:"mayTrustDevice,omitempty"`
	Methods        []struct {
		Type string `json:"type,omitempty"`
	} `json:"methods,omitempty"`
}

type AuthenticateTotpRequest struct {
	LoginId  string `json:"loginId,omitempty"`
	Remember bool   `json:"remember,omitempty"`
	Type     string `json:"type,omitempty"`
	Value    string `json:"value,omitempty"`
}

type AuthenticateResponse struct {
	AccessToken string `json:"access_token,omitempty"`
}

func sendAuthRequest(client *http.Client, data interface{}, targetIf interface{}) error {

	// TODO: for debug. remove me.
	func(v interface{}) {
		j, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}
		buf := bytes.NewBuffer(j)
		fmt.Printf("%v\n", buf.String())
	}(data)

	reqJson, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", authenticateURL, bytes.NewBuffer(reqJson))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://www.fastmail.com")
	req.Header.Set("Host", "www.fastmail.com")
	req.Header.Set("Accept-Language", "en-US,ja;q=0.7,en;q=0.3")
	req.Header.Set("Accept-Encoding", "deflate")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Accept", "application/json")

	req.Header.Set("Referer", "https://www.fastmail.com/login/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.15; rv:96.0) Gecko/20100101 Firefox/96.0")
	req.Header.Set("DNT", "1")
	req.Header.Set("TE", "trailers")

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(res.Body)
	if err != nil {
		return err
	}

	fmt.Printf("got: %v\n", buf.String())

	err = json.Unmarshal(buf.Bytes(), &targetIf)
	if err != nil {
		return err
	}

	return err
}

func Authenticate() {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	client := &http.Client{
		Jar: cookieJar,
	}

	res, err := client.Get(authenticateURL)
	if err != nil {
		panic(err)
	}

	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(res.Body)
	if err != nil {
		panic(err)
	}

	fmt.Printf("got: %v\n", buf.String())

	var authUsernameRes AuthenticateUsernameResponse
	if err := sendAuthRequest(client, AuthenticateUsernameRequest{Username: "mail@david.coffe"}, &authUsernameRes); err != nil {
		panic(err)
	}

	var authPassRes AuthenticatePasswordResponse
	if err := sendAuthRequest(client, AuthenticatePasswordRequest{
		LoginId:  authUsernameRes.LoginId,
		Remember: false,
		Type:     "password",
		Value:    "x",
	}, &authPassRes); err != nil {
		panic(err)
	}
}
