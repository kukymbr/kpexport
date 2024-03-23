package writer_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/kukymbr/kinopoiskexport/internal/app/kpvotes/domain"
	"github.com/kukymbr/kinopoiskexport/internal/app/kpvotes/writer"
	"github.com/kukymbr/kinopoiskexport/internal/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVotesIMDbCSVVotesWriter_WriteToFile(t *testing.T) {
	targetPath := "./testdata/target/test_imdb_csv_votes_writer.csv"

	_ = os.Remove(targetPath)
	t.Cleanup(func() {
		_ = os.Remove(targetPath)
	})

	now := time.Now()
	nowFmt := now.Format("2006-01-02")

	wr := writer.NewIMDbCSVVotesWriter(logger.NewDefaultConsoleLogger(true))
	votes := domain.Votes{
		domain.Vote{
			MovieNameOriginal: "Test Movie 1",
			Rate:              4,
			Timestamp:         now,
			MovieYear:         "2020",
		},
		domain.Vote{
			MovieNameRu: "Тест Фильм 2",
			Rate:        5,
			Timestamp:   now,
			MovieYear:   "2021",
		},
	}

	err := wr.WriteToFile(context.Background(), votes, targetPath)

	assert.NoError(t, err)
	assert.FileExists(t, targetPath)

	content, err := os.ReadFile(targetPath)
	require.NoError(t, err)

	assert.Contains(t, string(content), ",4,"+nowFmt+",Test Movie 1 (2020),")
	assert.Contains(t, string(content), ",5,"+nowFmt+",Тест Фильм 2 (2021),")
}
