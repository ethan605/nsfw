package crawler

// Session provides methods to access a crawling source
type Session interface {
	BaseURL() string
	FetchProfile(username string) string
	FetchOtherUsers(username string) string
}
