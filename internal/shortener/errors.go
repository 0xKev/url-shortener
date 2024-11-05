package shortener

import "fmt"

type InvalidURLError struct {
	ErrorMsg     string
	SubmittedURL string
}

func (i InvalidURLError) Error() string {
	return fmt.Sprintf("invalid url %s, %v", i.ErrorMsg, i.SubmittedURL)
}
