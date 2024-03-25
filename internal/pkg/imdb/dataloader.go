package imdb

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/kukymbr/kinopoiskexport/internal/pkg/downloader"
	"go.uber.org/zap"
)

const (
	TitleURL = "https://www.imdb.com/title/"
	findURL  = "https://www.imdb.com/find/"
)

func NewDataLoader(log *zap.Logger, downloader downloader.Downloader, cache Cache) DataLoader {
	return &dataLoader{
		log:        log,
		downloader: downloader,
		cache:      cache,
	}
}

type DataLoader interface {
	GetIDByTitle(ctx context.Context, title string) (TitleID, error)
}

type dataLoader struct {
	log        *zap.Logger
	downloader downloader.Downloader
	cache      Cache
}

func (d *dataLoader) GetIDByTitle(ctx context.Context, title string) (TitleID, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	if cached, err := d.cache.GetTitleID(ctx, title); err == nil && cached != "" {
		return cached, nil
	}

	query := url.Values{}
	query.Set("s", "all")
	query.Set("q", title)

	pageURL := findURL + "?" + query.Encode()

	body, err := d.downloader.Download(ctx, pageURL)
	if err != nil {
		return "", err
	}

	defer func() {
		_ = body.Close()
	}()

	doc, err := htmlquery.Parse(body)
	if err != nil {
		return "", fmt.Errorf("failed to parse body: %w", err)
	}

	item, err := htmlquery.Query(doc, `//a[@class="ipc-metadata-list-summary-item__t"]`)
	if err != nil {
		return "", fmt.Errorf("no item: %w", err)
	}

	if item == nil {
		return "", fmt.Errorf("no item")
	}

	href := htmlquery.SelectAttr(item, "href")
	if href == "" {
		return "", fmt.Errorf("no href")
	}

	href, ok := strings.CutPrefix(href, "/title/")
	if !ok {
		return "", fmt.Errorf("no title href")
	}

	parts := strings.SplitN(href, "/", 2)
	id := parts[0]

	if id == "" {
		return "", fmt.Errorf("no id")
	}

	tId := TitleID(id)

	_ = d.cache.StoreTitleID(ctx, title, tId)

	return tId, nil
}
