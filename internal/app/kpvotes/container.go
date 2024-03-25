package kpvotes

import (
	"context"
	"fmt"

	"github.com/kukymbr/godi"
	"github.com/kukymbr/kinopoiskexport/internal/app/kpvotes/reader"
	"github.com/kukymbr/kinopoiskexport/internal/app/kpvotes/writer"
	"github.com/kukymbr/kinopoiskexport/internal/pkg/downloader"
	"github.com/kukymbr/kinopoiskexport/internal/pkg/imdb"
	"github.com/kukymbr/kinopoiskexport/internal/pkg/kinopoisk"
	"go.uber.org/zap"
)

const (
	diLogger         = "logger"
	diDownloader     = "downloader"
	diImdbCache      = "imdb_cache"
	diImdbDataLoader = "imdb_dataloader"
	diVotesReader    = "votes_reader"
	diVotesWriter    = "votes_writer"
	diRunner         = "runner"
)

func buildContainer(ctx context.Context, log *zap.Logger, opt Options) (*godi.Container, error) {
	builder, err := getBuilder(ctx, log, opt)
	if err != nil {
		return nil, err
	}

	ctn, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build DI container: %w", err)
	}

	return ctn, nil
}

func getBuilder(ctx context.Context, log *zap.Logger, opt Options) (*godi.Builder, error) {
	builder := &godi.Builder{}

	err := builder.Add(
		godi.Def{
			Name: diLogger,
			Build: func(_ *godi.Container) (obj any, err error) {
				return log, nil
			},
		},
		godi.Def{
			Name: diDownloader,
			Build: func(ctn *godi.Container) (obj any, err error) {
				return downloader.NewStdDownloader(
					requireLogger(ctn),
					kinopoisk.TimeoutVotes,
					opt.ProxyURL,
				), nil
			},
		},
		godi.Def{
			Name: diImdbCache,
			Build: func(ctn *godi.Container) (obj any, err error) {
				cache := imdb.NewMemoryCache(requireLogger(ctn))

				if opt.IMDbCacheFile != "" {
					if err := cache.ImportTitlesIDs(ctx, opt.IMDbCacheFile); err != nil {
						return nil, err
					}
				}

				return cache, nil
			},
			Close: func(obj any) (err error) {
				cache := obj.(imdb.Cache)

				if opt.IMDbCacheFile != "" {
					_ = cache.ExportTitlesIDs(context.Background(), opt.IMDbCacheFile, true)
				}

				return nil
			},
		},
		godi.Def{
			Name: diImdbDataLoader,
			Build: func(ctn *godi.Container) (obj any, err error) {
				return imdb.NewDataLoader(
					requireLogger(ctn),
					requireDownloader(ctn),
					requireImdbCache(ctn),
				), nil
			},
		},
		godi.Def{
			Name: diVotesReader,
			Build: func(ctn *godi.Container) (obj any, err error) {
				return reader.NewVotesReader(
					requireLogger(ctn),
					requireDownloader(ctn),
					requireImdbDataLoader(ctn),
				), nil
			},
		},
		godi.Def{
			Name: diVotesWriter,
			Build: func(ctn *godi.Container) (obj any, err error) {
				return writer.NewIMDbCSVVotesWriter(requireLogger(ctn)), nil
			},
		},
		godi.Def{
			Name: diRunner,
			Build: func(ctn *godi.Container) (obj any, err error) {
				return &runner{
					log:    requireLogger(ctn),
					reader: requireReader(ctn),
					writer: requireWriter(ctn),
				}, nil
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to add dependencies to builder: %w", err)
	}

	return builder, nil
}

func requireLogger(ctn *godi.Container) *zap.Logger {
	return ctn.Get(diLogger).(*zap.Logger)
}

func requireDownloader(ctn *godi.Container) downloader.Downloader {
	return ctn.Get(diDownloader).(downloader.Downloader)
}

func requireImdbCache(ctn *godi.Container) imdb.Cache {
	return ctn.Get(diImdbCache).(imdb.Cache)
}

func requireImdbDataLoader(ctn *godi.Container) imdb.DataLoader {
	return ctn.Get(diImdbDataLoader).(imdb.DataLoader)
}

func requireReader(ctn *godi.Container) reader.VotesReader {
	return ctn.Get(diVotesReader).(reader.VotesReader)
}

func requireWriter(ctn *godi.Container) writer.VotesWriter {
	return ctn.Get(diVotesWriter).(writer.VotesWriter)
}

func requireRunner(ctn *godi.Container) *runner {
	return ctn.Get(diRunner).(*runner)
}
