package services

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"time"
)

const maxFetchedImageBytes = 5 << 20 // 5MB

var errBlockedAddress = errors.New("cards: refusing to fetch from a non-public address")

var baseDialer = &net.Dialer{Timeout: 5 * time.Second}

// safeDialContext resolves the host and validates the IP right before
// connecting — not just by inspecting the URL string — so a request can't
// reach an internal address via DNS rebinding (resolve safe at parse time,
// resolve differently at connect time) or a redirect to one. A person's
// photo URL is unauthenticated attacker-influenced input (submitted
// through the public sign-up form), so this is a real SSRF surface, not a
// theoretical one.
func safeDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
	if err != nil || len(ips) == 0 {
		return nil, errBlockedAddress
	}
	ip := ips[0]
	if !isPublicIP(ip) {
		return nil, errBlockedAddress
	}
	return baseDialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
}

func isPublicIP(ip net.IP) bool {
	return !ip.IsLoopback() && !ip.IsPrivate() && !ip.IsLinkLocalUnicast() &&
		!ip.IsLinkLocalMulticast() && !ip.IsUnspecified() && !ip.IsMulticast()
}

var safeImageClient = &http.Client{
	Timeout:   8 * time.Second,
	Transport: &http.Transport{DialContext: safeDialContext},
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 {
			return errors.New("cards: too many redirects")
		}
		return nil
	},
}

// fetchPersonPhoto best-effort fetches an attendee-supplied image URL to
// embed in their card. Any failure — network error, blocked/private
// address, non-image response, oversized body — just means the card
// renders without a photo; it never fails card generation over an
// untrusted third-party URL we don't control.
func fetchPersonPhoto(ctx context.Context, rawURL string) []byte {
	if rawURL == "" {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil || (req.URL.Scheme != "http" && req.URL.Scheme != "https") {
		return nil
	}

	res, err := safeImageClient.Do(req)
	if err != nil {
		return nil
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil
	}

	contentType := res.Header.Get("Content-Type")
	if contentType != "image/png" && contentType != "image/jpeg" {
		return nil
	}

	data, err := io.ReadAll(io.LimitReader(res.Body, maxFetchedImageBytes))
	if err != nil {
		return nil
	}
	return data
}
