package app

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	log15 "gopkg.in/inconshreveable/log15.v2"
)

// MakeAuthRequest makes http authorization request to Selectel Swift API
func MakeAuthRequest(user, pass, authURL string) (resp *http.Response, err error) {
	client := &http.Client{}
	//$ curl -i https://auth.selcdn.ru/ \
	r, err := http.NewRequest("GET", authURL, nil)
	if err != nil {
		return nil, errors.New("http.NewRequest error: " + err.Error())
	}
	//-H "X-Auth-User:user" \
	//-H "X-Auth-Key:password"
	r.Header.Add("X-Auth-User", user)
	r.Header.Add("X-Auth-Key", pass)
	log.Println(authURL+" Request: ", r)
	resp, err = client.Do(r)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, errors.New("http.Request.Do error: " + err.Error())
	}
	log.Println(authURL, " Response: ", resp)
	if resp.Header.Get("X-Auth-Token") != "" && resp.Header.Get("X-Expire-Auth-Token") != "" && resp.Header.Get("X-Storage-Token") != "" && resp.Header.Get("X-Storage-Url") != "" {
		return resp, nil
	}
	return nil, errors.New("makeAuthRequest error: Some of response headers are empty.")
}

// MakeStorageRequest makes http request to Selectel Swift API with given request method, URL,authToken and body
func MakeStorageRequest(method, storageURL, authToken string, reqBody io.Reader) ([]byte, error) {
	client := &http.Client{}
	r, err := http.NewRequest(method, storageURL, reqBody)
	if err != nil {
		return nil, errors.New("http.NewRequest error: " + err.Error())
	}
	r.Header.Add("X-Auth-Token", authToken)
	log15.Info("Request: ", "url", storageURL, "request", r)
	resp, err := client.Do(r)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, errors.New("http.Request.Do error: " + err.Error())
	}
	log15.Info(storageURL, " Response: ", resp)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("ioutil.ReadAll error: " + err.Error())
	}
	if resp.Status == "404" || strings.Contains(string(body), "Not Found") {
		return nil, errors.New("The file not found. ")
	}
	return body, nil
}
