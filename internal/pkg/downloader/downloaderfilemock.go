package downloader

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
)

func NewDownloaderFileMock(sources map[string]string) Downloader {
	return &downloaderFileMock{sources: sources}
}

type downloaderFileMock struct {
	sources map[string]string
}

func (d *downloaderFileMock) Download(_ context.Context, pageURL string) (body io.ReadCloser, err error) {
	errNotFound := fmt.Errorf("page %s not found in mock", pageURL)

	if d.sources == nil {
		return nil, errNotFound
	}

	file, ok := d.sources[pageURL]
	if !ok {
		return nil, errNotFound
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open mock file %s: %w", file, err)
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

func (d *downloaderFileMock) Close() error {
	return nil
}
