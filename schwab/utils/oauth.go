package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/samjtro/go-trade/utils"
	"github.com/spf13/viper"
)

func oAuthInit() TOKEN {
	var (
		m                   sync.Mutex
		tokens              TOKEN
		accessTokenResponse AccessTokenResponse
	)

	m.Lock()

	// oAuth Leg 1 - App Authorization
	openBrowser(fmt.Sprintf("https://api.schwabapi.com/v1/oauth/authorize?client_id=%s&redirect_uri=%s", viper.Get("APPKEY"), viper.Get("CBURL")))
	fmt.Printf("Log into your Schwab brokerage account. Copy Error404 URL and paste it here: ")
	var urlInput string
	fmt.Scanln(&urlInput)
	authCodeEncoded := getStringInBetween(urlInput, "?code=", "&session=")
	authCodeDecoded, err := url.QueryUnescape(authCodeEncoded)
	utils.Check(err)

	// oAuth Leg 2 - Access Token Creation
	authStringLegTwo := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", viper.Get("APPKEY"), viper.Get("SECRET")))))
	client := http.Client{}
	payload := fmt.Sprintf("grant_type=authorization_code&code=%s&redirect_uri=%s", authCodeDecoded, viper.Get("CBURL"))
	req, err := http.NewRequest("POST", "https://api.schwabapi.com/v1/oauth/token", bytes.NewBuffer([]byte(payload)))
	utils.Check(err)

	req.Header = http.Header{
		"Authorization": {authStringLegTwo},
		"Content-Type":  {"application/x-www-form-urlencoded"},
	}

	res, err := client.Do(req)
	utils.Check(err)
	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)
	utils.Check(err)

	err = json.Unmarshal(bodyBytes, &accessTokenResponse)
	utils.Check(err)

	tokens.Refresh = accessTokenResponse.refresh_token
	tokens.Bearer = accessTokenResponse.access_token
	tokens.BearerExpiration = time.Now().Add(time.Minute * 30)
	tokens.RefreshExpiration = time.Now().Add(time.Hour * 168)
	tokensJson, err := json.Marshal(tokens)
	utils.Check(err)

	f, err := os.Create("~/.foo/bar.json")
	utils.Check(err)
	defer f.Close()
	_, err = f.Write(tokensJson)
	utils.Check(err)

	m.Unlock()
	return tokens
}

func oAuthRefresh() string {
	var (
		m                   sync.Mutex
		accessTokenResponse AccessTokenResponse
	)

	m.Lock()
	tokens := readDB()

	authStringRefresh := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", viper.Get("APPKEY"), viper.Get("SECRET")))))
	client := http.Client{}
	req, err := http.NewRequest("POST", "https://api.schwabapi.com/v1/oauth/token", bytes.NewBuffer([]byte(fmt.Sprintf("grant_type=refresh_token&refresh_token=%s", tokens.Refresh))))
	utils.Check(err)

	req.Header = http.Header{
		"Authorization": {authStringRefresh},
		"Content-Type":  {"application/x-www-form-urlencoded"},
	}

	res, err := client.Do(req)
	utils.Check(err)

	err = json.NewDecoder(res.Body).Decode(&accessTokenResponse)
	utils.Check(err)

	m.Unlock()
	return accessTokenResponse.access_token
}
