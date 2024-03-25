package imdb

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sync"

	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
)

type ioTitleIDItem struct {
	Title string `json:"title,omitempty"`
	ID    string `json:"id,omitempty"`
}

func NewMemoryCache(log *zap.Logger) Cache {
	return &memoryCache{
		log: log.With(zap.String("who", "imdb.memoryCache")),
	}
}

type Cache interface {
	StoreTitleID(ctx context.Context, title string, id TitleID) error
	GetTitleID(ctx context.Context, title string) (TitleID, error)
	InvalidateTitleID(ctx context.Context, title string) error

	ExportTitlesIDs(ctx context.Context, targetPath string, append bool) error
	ImportTitlesIDs(ctx context.Context, sourcePath string) error
}

type memoryCache struct {
	log       *zap.Logger
	titlesIDs memoryTitleIDCache
}

type memoryTitleIDCache struct {
	sync.RWMutex
	items map[string]TitleID
}

func (m *memoryCache) StoreTitleID(ctx context.Context, title string, id TitleID) error {
	m.titlesIDs.Lock()
	defer m.titlesIDs.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}

	if m.titlesIDs.items == nil {
		m.titlesIDs.items = make(map[string]TitleID)
	}

	m.titlesIDs.items[title] = id

	return nil
}

func (m *memoryCache) GetTitleID(ctx context.Context, title string) (TitleID, error) {
	m.titlesIDs.RLock()
	defer m.titlesIDs.RUnlock()

	if err := ctx.Err(); err != nil {
		return "", err
	}

	if m.titlesIDs.items == nil {
		return "", nil
	}

	if id, ok := m.titlesIDs.items[title]; ok {
		return id, nil
	}

	return "", nil
}

func (m *memoryCache) InvalidateTitleID(ctx context.Context, title string) error {
	m.titlesIDs.Lock()
	defer m.titlesIDs.Unlock()

	if err := ctx.Err(); err != nil {
		return err
	}

	delete(m.titlesIDs.items, title)

	return nil
}

func (m *memoryCache) ExportTitlesIDs(ctx context.Context, targetPath string, append bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	log := m.log.Sugar().With("target", targetPath)

	log.Debug("Exporting cached")

	flags := os.O_RDWR | os.O_CREATE
	if append {
		flags |= os.O_APPEND
	} else {
		flags |= os.O_TRUNC
	}

	f, err := os.OpenFile(targetPath, flags, 0755)
	if err != nil {
		return fmt.Errorf("failed to create target cache file %s: %w", targetPath, err)
	}

	defer func() {
		_ = f.Close()
	}()

	m.titlesIDs.RLock()
	defer m.titlesIDs.RUnlock()

	log.Debugf("Exporting %d item(s)", len(m.titlesIDs.items))

	if len(m.titlesIDs.items) == 0 {
		return nil
	}

	written := 0

	for title, id := range m.titlesIDs.items {
		if err := ctx.Err(); err != nil {
			return err
		}

		item := ioTitleIDItem{
			Title: title,
			ID:    id.String(),
		}

		row, err := jsoniter.MarshalToString(item)
		if err != nil {
			log.Warnf("Failed to marshal row (title=%s, id=%s): %s", title, id, err)

			continue
		}

		if _, err := f.WriteString(row + "\n"); err != nil {
			log.Warnf("Failed to write row (title=%s, id=%s): %s", title, id, err)

			continue
		}

		written++
	}

	log.Debugf("Exported %d item(s)", written)

	return nil
}

// TODO: import from directory
func (m *memoryCache) ImportTitlesIDs(ctx context.Context, sourcePath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	log := m.log.Sugar().With("source", sourcePath)

	log.Debug("Importing cached")

	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		log.Debug("No file exist")

		return nil
	}

	f, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source cache file %s: %w", sourcePath, err)
	}

	defer func() {
		_ = f.Close()
	}()

	m.titlesIDs.Lock()
	defer m.titlesIDs.Unlock()

	if m.titlesIDs.items == nil {
		m.titlesIDs.items = make(map[string]TitleID)
	}

	scanner := bufio.NewScanner(f)

	imported := 0

	for scanner.Scan() {
		var item ioTitleIDItem

		if err := jsoniter.Unmarshal(scanner.Bytes(), &item); err != nil {
			log.Warnf("Failed to unmarshal row: %s", err)

			continue
		}

		if item.Title == "" || item.ID == "" {
			log.Debugf("Item is skipped (title=%s, id=%s)", item.Title, item.ID)
		}

		m.titlesIDs.items[item.Title] = TitleID(item.ID)

		imported++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan %s: %w", sourcePath, err)
	}

	log.Debugf("Imported %d items", imported)

	return nil
}
