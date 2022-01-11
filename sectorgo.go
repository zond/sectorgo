package sectorgo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"

	"golang.org/x/net/html"
)

const (
	loginPage                    = "https://mypagesapi.sectoralarm.net/User/Login"
	statusPage                   = "https://mypagesapi.sectoralarm.net/Panel/GetOverview/"
	requestVerificationTokenName = "__RequestVerificationToken"
)

var (
	versionTokenReg = regexp.MustCompile("^/Scripts/main\\.js\\?(.*)$")
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
	verificationToken := ""
	versionToken := ""
	tokenizer := html.NewTokenizer(resp.Body)
	for tokenType := tokenizer.Next(); tokenType != html.ErrorToken; tokenType = tokenizer.Next() {
		token := tokenizer.Token()
		switch token.Data {
		case "input":
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
				verificationToken = value
			}
		case "script":
			for _, attr := range token.Attr {
				if attr.Key == "src" {
					if match := versionTokenReg.FindStringSubmatch(attr.Val); match != nil {
						versionToken = match[1]
					}
				}
			}
		}
	}
	if verificationToken == "" {
		return nil, fmt.Errorf("Found no %q in %q", requestVerificationTokenName, loginPage)
	}
	if versionToken == "" {
		return nil, fmt.Errorf("Found no %v in %q", versionTokenReg, loginPage)
	}
	_, err = client.PostForm(loginPage, url.Values{"userID": {userID}, "password": {password}})
	if err != nil {
		return nil, err
	}
	resp, err = client.PostForm(statusPage, url.Values{"Version": {versionToken}})
	if err != nil {
		return nil, err
	}
	status := &Status{}
	if err := json.NewDecoder(resp.Body).Decode(status); err != nil {
		return nil, err
	}
	return status, nil
}
