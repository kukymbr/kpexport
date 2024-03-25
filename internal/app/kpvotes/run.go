package kpvotes

import (
	"context"
	"fmt"

	"github.com/kukymbr/kinopoiskexport/internal/app/kpvotes/reader"
	"github.com/kukymbr/kinopoiskexport/internal/app/kpvotes/writer"
	"go.uber.org/zap"
)

func Run(ctx context.Context, log *zap.Logger, opt Options) error {
	if err := opt.SetFromEnv(); err != nil {
		return fmt.Errorf("failed to set options from environment variables: %w", err)
	}

	ctn, err := buildContainer(ctx, log, opt)
	if err != nil {
		return err
	}

	defer func() {
		_ = ctn.Close()
	}()

	runner := requireRunner(ctn)

	return runner.Run(ctx, opt)
}

type runner struct {
	log    *zap.Logger
	reader reader.VotesReader
	writer writer.VotesWriter
}

func (r *runner) Run(ctx context.Context, opt Options) error {
	log := r.log.With(zap.String("who", "runner"), zap.String("uid", opt.UserID.String()))

	log.Info("Reading votes")

	votes, err := r.reader.ReadVotes(ctx, opt.UserID)
	if err != nil {
		return fmt.Errorf("failed to read votes: %w", err)
	}

	log.Info("Writing votes")

	if err := r.writer.WriteToFile(ctx, votes, opt.TargetFile); err != nil {
		return fmt.Errorf("failed to write votes: %w", err)
	}

	log.Info("Votes written to the " + opt.TargetFile)

	return nil
}
