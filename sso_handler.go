package main

import (
	"context"
	"io"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/random"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var ssoConfig *oauth2.Config
var stateValue string

const gUserInfoUrl = "https://www.googleapis.com/oauth2/v2/userinfo"
const gUserInfoScope = "https://www.googleapis.com/auth/userinfo.email"

func AddSsoHandlers(e *echo.Echo) {
	e.GET("/auth", loginHandler)
	e.GET("/loggedin", ssoCallbackHandler)

	stateValue = random.String(20)

	ssoConfig = &oauth2.Config{
		RedirectURL:  os.Getenv("SERVER_URL") + "/loggedin",
		ClientID:     os.Getenv("G_AUTH_ID"),
		ClientSecret: os.Getenv("G_AUTH_SECRET"),
		Scopes:       []string{gUserInfoScope},
		Endpoint:     google.Endpoint,
	}
}

func loginHandler(c echo.Context) error {
	url := ssoConfig.AuthCodeURL(stateValue)
	logs.info("Redirecting to Google SSO...")
	return c.Redirect(http.StatusTemporaryRedirect, url)
}

func ssoCallbackHandler(c echo.Context) error {
	logs.info("Callback from Google SSO...")
	getUserData(c.FormValue("code"))
	return c.Redirect(http.StatusTemporaryRedirect, "/")
}

func getUserData(code string) ([]byte, error) {
	token, err := ssoConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, err
	}
	response, err := http.Get(gUserInfoUrl + "?access_token=" + token.AccessToken)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	logs.info("Data from Google Auth = %s", data)
	return data, nil
}
