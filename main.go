package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/garyburd/go-oauth/oauth"
)

var authorizationClient = oauth.Client{
	TemporaryCredentialRequestURI: "https://www.hatena.com/oauth/initiate",
	ResourceOwnerAuthorizationURI: "https://www.hatena.ne.jp/oauth/authorize",
	TokenRequestURI:               "https://www.hatena.com/oauth/token",
}

var credentialFile = flag.String("config", "config.json", "Oauth 1.0a Credential File.")
var listFile = flag.String("list", "bookmark.list", "bookmark list")

func readCredentials() error {
	binary, err := ioutil.ReadFile(*credentialFile)
	if err != nil {
		return err
	}

	return json.Unmarshal(binary, &authorizationClient.Credentials)
}

func readList() (string, error) {
	binary, err := ioutil.ReadFile(*listFile)
	if err != nil {
		return "", err
	}

	return string(binary), err
}

func getOAuthToken() (*oauth.Credentials, error) {
	err := readCredentials()
	if err != nil {
		return nil, err
	}

	scope := url.Values{
		"scope": {"read_public,write_public"},
	}
	tempCredeintal, err := authorizationClient.RequestTemporaryCredentials(http.DefaultClient, "oob", scope)
	if err != nil {
		return nil, err
	}

	authURL := authorizationClient.AuthorizationURL(tempCredeintal, nil)

	fmt.Printf("Access to %s\nInput Token: ", authURL)
	var code string
	fmt.Scanln(&code)

	token, _, err := authorizationClient.RequestToken(http.DefaultClient, tempCredeintal, code)
	return token, err
}

func getBookmark(bookmarkURL string, token *oauth.Credentials) error {
	apiURL := "http://api.b.hatena.ne.jp/1/my/bookmark"
	query := url.Values{
		"url": {bookmarkURL},
	}
	res, err := authorizationClient.Get(http.DefaultClient, token, apiURL, query)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	fmt.Println(res.Status)
	if _, err := io.Copy(os.Stdout, res.Body); err != nil {
		return err
	}

	return err
}

func setBookmark(bookmarkURL string, token *oauth.Credentials) error {
	apiURL := "http://api.b.hatena.ne.jp/1/my/bookmark"
	query := url.Values{
		"url": {bookmarkURL},
	}
	res, err := authorizationClient.Post(http.DefaultClient, token, apiURL, query)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	fmt.Println(res.Status)
	if _, err := io.Copy(os.Stdout, res.Body); err != nil {
		return err
	}
	return err
}

func main() {
	if err := realMain(); err != nil {
		log.Fatal(err)
	}
}

func realMain() error {
	token, err := getOAuthToken()
	if err != nil {
		return err
	}

	list, err := readList()
	if err != nil {
		return err
	}

	for i, line := range strings.Split(list, "\n") {
		if line == "" {
			continue
		}
		fmt.Printf("\nNumber: %00d\n", i)

		err = setBookmark(line, token)
		if err != nil {
			return err
		}
		time.Sleep(350000000)
	}

	return nil
}
