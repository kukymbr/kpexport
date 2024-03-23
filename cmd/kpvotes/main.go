package main

import (
	"context"

	"github.com/kukymbr/kinopoiskexport/internal/app/kpvotes"
	"github.com/kukymbr/kinopoiskexport/internal/pkg/logger"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var log *zap.Logger

var opt = kpvotes.Options{
	IsDebug: false,
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cmd := initCommand(ctx)

	// TODO: listen to signals

	if err := cmd.ExecuteContext(ctx); err != nil {
		log.Sugar().Fatalf("failed with error: %s", err)
	}

	log.Info("Done.")
}

func initCommand(ctx context.Context) *cobra.Command {
	root := &cobra.Command{
		Use:   "kpvotes",
		Short: "Export votes in CSV",
		Long: "Export the user's movies votes from the kinopoisk.ru into a CSV file. " +
			"The environment variables are acceptable: \n" +
			"- KPEXPORT_PROXY_URL: downloader client proxy URL",

		SilenceErrors: true,
		SilenceUsage:  true,

		RunE: func(cmd *cobra.Command, args []string) error {
			initLogger()

			return kpvotes.Run(ctx, log, opt)
		},
	}

	root.Flags().StringVar(&opt.TargetFile, "target", "", "target .csv file path")
	root.Flags().BoolVar(&opt.IsDebug, "debug", false, "enable the debug mode")
	root.Flags().Var(&opt.UserID, "uid", "kinopoisk user ID")

	_ = root.MarkFlagRequired("target")
	_ = root.MarkFlagRequired("uid")

	return root
}

func initLogger() {
	log = logger.NewDefaultConsoleLogger(opt.IsDebug)
}
