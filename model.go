package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type SteamResponse struct {
	Steamid string
	Success int
}

type Steam struct {
	Response SteamResponse
}

type Profile struct {
	Personaname string
	Avatarfull  string
}

type DotaProfile struct {
	Profile Profile
}

func getData(url string) []byte {

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println(string(body))
	return body
}

func Rawurlencode(str string) string {
	return strings.Replace(url.QueryEscape(str), "+", "%20", -1)
}
