package crawler

// Session provides methods to access a crawling source
type Session interface {
	BaseURL() string
	ProfileURL(username string) string
}
