// Package oauth wraps Google's OAuth2 authorization-code flow, the same
// shape as internal/pkg/storage (Connect) and internal/pkg/mailer (New):
// one small package around a third-party SDK, constructed once in
// cmd/api/main.go and injected into whichever module needs it. *oauth2.Token
// never leaves this package — see ExchangeAndFetchUserInfo.
package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"identitycard-server/internal/config"
)

// ErrNotConfigured is returned by every method when the app was booted
// without Google credentials — a deliberate placeholder state (see
// config.Config.GoogleClientID) rather than a startup failure.
var ErrNotConfigured = errors.New("oauth: google client id/secret/redirect url not configured")

type UserInfo struct {
	Sub      string `json:"sub"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Picture  string `json:"picture"`
	Verified bool   `json:"email_verified"`
}

type Client struct {
	conf *oauth2.Config
}

func New(cfg *config.Config) *Client {
	if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" || cfg.GoogleRedirectURL == "" {
		return &Client{}
	}
	return &Client{
		conf: &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
	}
}

func (c *Client) Configured() bool { return c.conf != nil }

// AuthURL builds the Google consent-screen URL for the given opaque state
// (see auth.SignOAuthState — it carries intent/plan_code/organization_name
// through the redirect round trip).
func (c *Client) AuthURL(state string) (string, error) {
	if !c.Configured() {
		return "", ErrNotConfigured
	}
	return c.conf.AuthCodeURL(state, oauth2.AccessTypeOnline), nil
}

// ExchangeAndFetchUserInfo exchanges the authorization code for a token and
// immediately fetches the Google profile it authorizes, so the raw
// *oauth2.Token never has to leave this package — callers only ever see
// UserInfo.
func (c *Client) ExchangeAndFetchUserInfo(ctx context.Context, code string) (UserInfo, error) {
	if !c.Configured() {
		return UserInfo{}, ErrNotConfigured
	}
	token, err := c.conf.Exchange(ctx, code)
	if err != nil {
		return UserInfo{}, err
	}

	httpClient := c.conf.Client(ctx, token)
	resp, err := httpClient.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		return UserInfo{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return UserInfo{}, errors.New("oauth: userinfo request failed")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return UserInfo{}, err
	}

	var info UserInfo
	if err := json.Unmarshal(body, &info); err != nil {
		return UserInfo{}, err
	}
	return info, nil
}
