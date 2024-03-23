package kinopoisk

import (
	"fmt"
	"strings"
	"time"
)

const (
	Host = "https://www.kinopoisk.ru"

	TimeoutVotes = 60 * time.Second
)

func ValidateKpURL(pageURL string) error {
	if !strings.HasPrefix(strings.ToLower(pageURL), Host+"/") {
		return fmt.Errorf("the '%s' URL is not a valid kinopois address", pageURL)
	}

	return nil
}
