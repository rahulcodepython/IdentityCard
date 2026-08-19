package entities

// MessageResponse is the shared shape for every endpoint whose only
// response is a human-readable confirmation — one definition, reused by
// every domain, instead of a MessageResponse-per-domain that would all be
// byte-for-byte identical.
type MessageResponse struct {
	Message string `json:"message"`
}
