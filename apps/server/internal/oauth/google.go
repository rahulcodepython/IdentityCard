// Package oauth wraps Google's OAuth2 authorization-code flow, the same
// shape as internal/storage (Connect) and internal/mailer (New): one small
// package around a third-party SDK, constructed once in cmd/api/main.go
// and injected into whichever module needs it (internal/modules/auth).
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

func (c *Client) Exchange(ctx context.Context, code string) (*oauth2.Token, error) {
	if !c.Configured() {
		return nil, ErrNotConfigured
	}
	return c.conf.Exchange(ctx, code)
}

// FetchUserInfo calls Google's OpenID userinfo endpoint with the access
// token from Exchange.
func (c *Client) FetchUserInfo(ctx context.Context, token *oauth2.Token) (UserInfo, error) {
	if !c.Configured() {
		return UserInfo{}, ErrNotConfigured
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
