package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type AuthentikAuth struct {
	url string

	clientID string
	username string
	password string
}

func NewAuthentikAuth(url, clientID, username, password string) *AuthentikAuth {
	return &AuthentikAuth{
		url:      url,
		clientID: clientID,
		username: username,
		password: password,
	}
}

func (a AuthentikAuth) Authenticate() (string, error) {

	// gather data
	endpoint := fmt.Sprintf("%s/application/o/token/", a.url)
	payload := url.Values{
		"grant_type": {"client_credentials"},
		"client_id":  {a.clientID},
		"username":   {a.username},
		"password":   {a.password},
	}

	// send request
	resp, err := http.PostForm(endpoint, payload)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// read response
	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			errMsg = fmt.Errorf("%s: %s", errMsg, string(body))
		}
		return "", errMsg
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}
	accessToken := data["access_token"].(string)

	return accessToken, nil
}
