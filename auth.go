package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
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

type AuthenticateResponse struct {
	AccountType     string            `json:"accountType,omitempty"`
	SigningId       string            `json:"signingId,omitempty"`
	SigningKey      string            `json:"signingKey,omitempty"`
	IsAdmin         bool              `json:"isAdmin,omitempty"`
	SessionKey      string            `json:"sessionKey,omitempty"`
	PrimaryAccounts map[string]string `json:"primaryAccounts,omitempty"`
	AccessToken     string            `json:"accessToken,omitempty"`
	ApiUrl          string            `json:"apiUrl,omitempty"`
	UserID          string            `json:"userId,omitempty"`
}

type AuthenticateTotpRequest struct {
	LoginId  string `json:"loginId,omitempty"`
	Remember bool   `json:"remember,omitempty"`
	Type     string `json:"type,omitempty"`
	Value    string `json:"value,omitempty"`
}

func sendAuthRequest(client *http.Client, data interface{}, targetIf interface{}) (*bytes.Buffer, error) {
	reqJson, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", authenticateURL, bytes.NewBuffer(reqJson))
	if err != nil {
		return nil, err
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
		return nil, err
	}

	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(res.Body)
	if err != nil {
		return nil, err
	}

	if targetIf != nil {
		err = json.Unmarshal(buf.Bytes(), &targetIf)
		if err != nil {
			return nil, err
		}
	}

	return buf, err
}

func Authenticate(username, password string) (*AuthenticateResponse, error) {
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Jar: cookieJar,
	}

	res, err := client.Get(authenticateURL)
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(res.Body)
	if err != nil {
		return nil, err
	}

	var authUsernameRes AuthenticateUsernameResponse
	if _, err := sendAuthRequest(client, AuthenticateUsernameRequest{Username: username}, &authUsernameRes); err != nil {
		return nil, err
	}

	var authPassRes AuthenticateResponse
	buf, err = sendAuthRequest(client, AuthenticatePasswordRequest{
		LoginId:  authUsernameRes.LoginId,
		Remember: true,
		Type:     "password",
		Value:    password,
	}, &authPassRes)

	if err != nil {
		return nil, err
	}

	// check if we got a access token
	if authPassRes.AccessToken != "" {
		return &authPassRes, nil
	}

	// try to parse into other struct
	var totpReq AuthenticatePasswordResponse
	err = json.Unmarshal(buf.Bytes(), &totpReq)
	if err != nil {
		return nil, err
	}

	if len(totpReq.Methods) > 0 {
		foundTotp := false
		for _, method := range totpReq.Methods {
			if method.Type == "totp" {
				foundTotp = true
				break
			}
		}

		if foundTotp {
			fmt.Print("enter your 2fa token: ")
			input := bufio.NewScanner(os.Stdin)
			input.Scan()

			var authRes AuthenticateResponse
			if _, err := sendAuthRequest(client, AuthenticateTotpRequest{
				LoginId:  authUsernameRes.LoginId,
				Remember: true,
				Type:     "totp",
				Value:    input.Text(),
			}, &authRes); err != nil {
				return nil, err
			}

			return &authRes, nil
		}
	}

	return nil, errors.New("login failed")
}
