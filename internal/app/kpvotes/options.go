package kpvotes

import (
	"fmt"
	"net/url"
	"os"

	"github.com/kukymbr/kinopoiskexport/internal/pkg/kinopoisk"
)

const (
	envPrefix   = "KPEXPORT_"
	envProxyURL = envPrefix + "PROXY_URL"
)

type Options struct {
	UserID   kinopoisk.UserID
	ProxyURL *url.URL

	TargetFile    string
	IMDbCacheFile string

	TargetChunkSize uint

	IsDebug bool
}

func (o *Options) SetFromEnv() error {
	if val := os.Getenv(envProxyURL); val != "" {
		proxy, err := url.Parse(val)
		if err != nil {
			return fmt.Errorf("failed to parse proxy URL from the env var %s: %w", envProxyURL, err)
		}

		o.ProxyURL = proxy
	}

	return nil
}
