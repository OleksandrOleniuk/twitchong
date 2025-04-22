package shared

// OAuthTokenChan is a channel used to pass the OAuth token from the callback handler to the main goroutine
var OAuthTokenChan = make(chan string, 1) 