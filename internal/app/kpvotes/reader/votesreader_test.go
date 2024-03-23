package reader_test

import (
	"context"
	"testing"

	"github.com/kukymbr/kinopoiskexport/internal/app/kpvotes/reader"
	"github.com/kukymbr/kinopoiskexport/internal/pkg/downloader"
	"github.com/kukymbr/kinopoiskexport/internal/pkg/imdb"
	"github.com/kukymbr/kinopoiskexport/internal/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestVotesReader_ReadVotes(t *testing.T) {
	log := logger.NewDefaultConsoleLogger(true)
	dwn := downloader.NewDownloaderFileMock(map[string]string{
		"https://www.kinopoisk.ru/user/33666291/votes/list/vs/vote/perpage/200/page/1": "./testdata/votes_page1.html",
		"https://www.kinopoisk.ru/user/33666291/votes/list/vs/vote/perpage/200/page/2": "./testdata/votes_page2.html",
		"https://www.imdb.com/find/?q=Anatomie+d%27une+chute+%282023%29&s=all":         "./testdata/imdb_1.html",
	})
	imdbDL := imdb.NewDataLoader(log, dwn)
	rd := reader.NewVotesReader(log, dwn, imdbDL)

	votes, err := rd.ReadVotes(context.Background(), 33666291)

	assert.NoError(t, err)
	assert.Len(t, votes, 1)
}
