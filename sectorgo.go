package sectorgo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"golang.org/x/net/html"
)

const (
	loginPage                    = "https://mypagesapi.sectoralarm.net/User/Login"
	statusPage                   = "https://mypagesapi.sectoralarm.net/Panel/GetOverview/"
	requestVerificationTokenName = "__RequestVerificationToken"
)

type Panel struct {
	ArmedStatus string
}

type Status struct {
	Panel Panel
}

func GetStatus(userID, password string) (*Status, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{
		Jar: jar,
	}
	resp, err := client.Get(loginPage)
	if err != nil {
		return nil, err
	}
	verToken := ""
	tokenizer := html.NewTokenizer(resp.Body)
	for tokenType := tokenizer.Next(); tokenType != html.ErrorToken; tokenType = tokenizer.Next() {
		token := tokenizer.Token()
		if token.Data == "input" {
			isVerToken := false
			value := ""
			for _, attr := range token.Attr {
				if attr.Key == "name" && attr.Val == requestVerificationTokenName {
					isVerToken = true
				}
				if attr.Key == "value" {
					value = attr.Val
				}
			}
			if isVerToken {
				verToken = value
				break
			}
		}
	}
	if verToken == "" {
		return nil, fmt.Errorf("Found no %q in %q", requestVerificationTokenName, loginPage)
	}
	_, err = client.PostForm(loginPage, url.Values{"userID": {userID}, "password": {password}})
	if err != nil {
		return nil, err
	}
	resp, err = client.Post(statusPage, "", &bytes.Buffer{})
	if err != nil {
		return nil, err
	}
	status := &Status{}
	if err := json.NewDecoder(resp.Body).Decode(status); err != nil {
		return nil, err
	}
	return status, nil
}
