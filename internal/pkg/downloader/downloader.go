package downloader

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
)

func NewStdDownloader(log *zap.Logger, timeout time.Duration, proxyURL *url.URL) Downloader {
	client := &http.Client{
		Timeout: timeout,
	}

	if proxyURL != nil {
		client.Transport = &http.Transport{
			Proxy:           http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	return NewStdDownloaderWithClient(log, client)
}

func NewStdDownloaderWithClient(log *zap.Logger, httpClient *http.Client) Downloader {
	return &stdDownloader{
		log:    log.With(zap.String("who", "stdDownloader")),
		client: httpClient,
	}
}

// Downloader is a tool to download page's HTML content.
type Downloader interface {
	// Download downloads the content.
	Download(ctx context.Context, pageURL string) (body io.ReadCloser, err error)

	io.Closer
}

type stdDownloader struct {
	log    *zap.Logger
	client *http.Client
}

func (d *stdDownloader) Download(ctx context.Context, pageURL string) (body io.ReadCloser, err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	log := d.log.With(zap.String("page_url", pageURL))

	if _, err := url.Parse(pageURL); err != nil {
		return nil, fmt.Errorf("page URL '%s' is invalid: %w", pageURL, err)
	}

	log.Debug("Creating request")

	req, err := http.NewRequest(http.MethodGet, pageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request to %s: %w", pageURL, err)
	}

	req.Header.Set("Accept", "text/html")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9")

	log.Debug("Sending request")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request to '%s' failed: %w", pageURL, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got non-OK response from '%s': %d", pageURL, resp.StatusCode)
	}

	log.Debug("Downloaded")

	return resp.Body, nil
}

func (d *stdDownloader) Close() error {
	if d.client != nil {
		d.client.CloseIdleConnections()
	}

	return nil
}
