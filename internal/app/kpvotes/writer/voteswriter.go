package writer

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kukymbr/kinopoiskexport/internal/pkg/kinopoisk"
	"github.com/kukymbr/kinopoiskexport/internal/pkg/utils"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func NewIMDbCSVVotesWriter(log *zap.Logger) VotesWriter {
	return &votesIMDbCSVVotesWriter{
		log: log.With(zap.String("who", "votesIMDbCSVVotesWriter")),
	}
}

type VotesWriter interface {
	WriteToFile(ctx context.Context, votes kinopoisk.Votes, targetPath string, chunkSize uint) error
}

type votesIMDbCSVVotesWriter struct {
	log *zap.Logger
}

func (v *votesIMDbCSVVotesWriter) WriteToFile(
	ctx context.Context,
	votes kinopoisk.Votes,
	targetPath string,
	chunkSize uint,
) error {
	if chunkSize == 0 {
		return v.writeToFile(ctx, votes, targetPath)
	}

	chunks := utils.Chunk(votes, chunkSize)

	targetPath = utils.FixSeparators(targetPath)
	dir := filepath.Dir(targetPath)
	filename := filepath.Base(targetPath)
	ext := filepath.Ext(filename)

	if ext != "" {
		filename, _ = strings.CutSuffix(filename, ext)
	}

	getTargetPath := func(chunkN int) string {
		return filepath.Join(dir, filename+"."+fmt.Sprintf("%d", chunkN)+ext)
	}

	eg, ctx := errgroup.WithContext(ctx)

	for i, votes := range chunks {
		chunkN := i
		votes := votes

		eg.Go(func() error {
			return v.writeToFile(ctx, votes, getTargetPath(chunkN))
		})
	}

	return eg.Wait()
}

func (v *votesIMDbCSVVotesWriter) writeToFile(
	ctx context.Context,
	votes kinopoisk.Votes,
	targetPath string,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	log := v.log.With(zap.String("target_path", targetPath)).Sugar()

	log.Info("Creating file")

	f, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create target file %s: %w", targetPath, err)
	}

	defer func() {
		_ = f.Close()
	}()

	header := []string{
		"Const",
		"Your Rating",
		"Date Rated",
		"Title",
		"URL",
		"Title Type", "IMDb Rating", "Runtime (mins)",
		"Year", "Genres", "Num Votes", "Release Date", "Directors",
	}

	writer := csv.NewWriter(f)
	defer writer.Flush()

	log.Info("Writing header")

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	for i, vote := range votes {
		log.Debugf("Writing row #%d", i)

		row := []string{
			vote.ImdbID.String(),
			fmt.Sprintf("%d", vote.Rate),
			vote.Timestamp.Format("2006-01-02"),
			vote.GetOriginalTitle(),
			vote.ImdbID.ToURL(),
			"", "", "", "",
			"", "", "", "", "",
		}

		if err := writer.Write(row); err != nil {
			log.Warnf("Failed to write row #%d: %s", i, err)

			continue
		}
	}

	return nil
}
