package reader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/kukymbr/kinopoiskexport/internal/app/kpvotes/domain"
	"github.com/kukymbr/kinopoiskexport/internal/pkg/downloader"
	"github.com/kukymbr/kinopoiskexport/internal/pkg/imdb"
	"github.com/kukymbr/kinopoiskexport/internal/pkg/kinopoisk"
	"go.uber.org/zap"
	"golang.org/x/net/html"
)

var errNothingFound = errors.New("nothing found")

func NewVotesReader(log *zap.Logger, downloader downloader.Downloader, imdbLoader imdb.DataLoader) VotesReader {
	return &votesReader{
		log:            log.With(zap.String("who", "votesReader")),
		downloader:     downloader,
		imdbDataLoader: imdbLoader,
	}
}

type VotesReader interface {
	ReadVotes(ctx context.Context, userID kinopoisk.UserID) (domain.Votes, error)
}

type votesReader struct {
	log            *zap.Logger
	downloader     downloader.Downloader
	imdbDataLoader imdb.DataLoader
}

func (r *votesReader) ReadVotes(ctx context.Context, userID kinopoisk.UserID) (domain.Votes, error) {
	log := r.log.With(zap.String("uid", userID.String()))
	pageN := uint16(1)
	votes := make(domain.Votes, 0)

	// TODO: download pages in several goroutines
	for {
		pageVotes, err := r.readPage(ctx, log, userID, pageN)

		if err == nil && len(pageVotes) == 0 || errors.Is(err, errNothingFound) {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("failed to read votes page #%d for user %s: %w", pageN, userID.String(), err)
		}

		votes = append(votes, pageVotes...)

		pageN++
	}

	return votes, nil
}

func (r *votesReader) readPage(
	ctx context.Context,
	log *zap.Logger,
	userID kinopoisk.UserID,
	pageN uint16,
) (domain.Votes, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	log = r.log.With(zap.Uint16("page", pageN))

	log.Debug("Reading votes")

	pageURL := userID.ToURL() + "/votes/list/vs/vote/perpage/200/page/" + fmt.Sprintf("%d", pageN)

	body, err := r.downloader.Download(ctx, pageURL)
	if err != nil {
		log.Error("failed to download: " + err.Error())

		return nil, err
	}

	defer func() {
		_ = body.Close()
	}()

	return r.parseHTML(ctx, log, body, pageN)
}

func (r *votesReader) parseHTML(
	ctx context.Context,
	log *zap.Logger,
	body io.Reader,
	pageN uint16,
) (domain.Votes, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	doc, err := htmlquery.Parse(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse body: %w", err)
	}

	errBox, _ := htmlquery.Query(doc, `//form[@id="f_filtr"]`)
	if errBox != nil {
		if strings.Contains(htmlquery.InnerText(errBox), "Ни одной записи не найдено") {
			return nil, errNothingFound
		}
	}

	itemNodes, err := htmlquery.QueryAll(doc, `//div[@class="profileFilmsList"]/div[`+xpathClass("item")+`]`)
	if err != nil {
		return nil, fmt.Errorf("failed to parse items: %w", err)
	}

	votes := make(domain.Votes, 0, len(itemNodes))

	for i, node := range itemNodes {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		vote := r.parseItemNode(log, node)

		if vote != nil {
			imdbID, err := r.imdbDataLoader.GetIDByTitle(ctx, vote.GetOriginalTitle())
			if err != nil {
				log.Debug("failed to get IMDb ID for " + vote.GetOriginalTitle() + ": " + err.Error())

				continue
			}

			vote.ImdbID = imdbID

			if err := votes.Add(*vote); err != nil {
				log.Debug(err.Error())

				continue
			}
		}

		log.Info(fmt.Sprintf("[read][page#%d][%03d/%03d] done", pageN, i+1, len(itemNodes)))
	}

	return votes, nil
}

func (r *votesReader) parseItemNode(log *zap.Logger, node *html.Node) *domain.Vote {
	nameRusNode, err := htmlquery.Query(node, `//div[@class="nameRus"]/a[@href]`)
	if err != nil {
		log.Warn("no .nameRus: " + err.Error())

		return nil
	}

	if nameRusNode == nil {
		log.Warn("no .nameRus found")

		return nil
	}

	vote := domain.Vote{}
	vote.MovieURL = htmlquery.SelectAttr(nameRusNode, "href")
	vote.MovieNameRu, vote.MovieYear = parseRuName(htmlquery.InnerText(nameRusNode))

	if nameOrigNode, err := htmlquery.Query(node, `//div[@class="nameEng"]`); err == nil {
		vote.MovieNameOriginal = getInnerText(nameOrigNode)
	}

	if dateNode, err := htmlquery.Query(node, `div[@class="date"]`); err == nil {
		dateText := getInnerText(dateNode)

		if timestamp, err := time.Parse("02.01.2006, 15:04", dateText); err == nil {
			vote.Timestamp = timestamp
		} else {
			vote.Timestamp = time.Now()
		}
	}

	if voteNode, err := htmlquery.Query(node, `div[@class="vote"]`); err == nil {
		voteStr := getInnerText(voteNode)

		voteVal, err := strconv.ParseUint(voteStr, 10, 8)
		if err == nil {
			vote.Rate = uint8(voteVal)
		}
	}

	return &vote
}

func (r *votesReader) processParsedVote(ctx context.Context, vote *domain.Vote) {
	imdbID, err := r.imdbDataLoader.GetIDByTitle(ctx, vote.GetOriginalTitle())
	if err != nil {
		r.log.Debug("failed to get IMDb ID for " + vote.GetOriginalTitle() + ": " + err.Error())

		return
	}

	vote.ImdbID = imdbID
}

func xpathClass(class string) string {
	return `contains(concat(" ", normalize-space(@class), " "), " ` + class + ` ")`
}

func getInnerText(node *html.Node) string {
	if node == nil {
		return ""
	}

	text := htmlquery.InnerText(node)

	text = strings.Replace(text, "\n", " ", -1)
	text = strings.Replace(text, `"`, `'`, -1)

	return text
}

func parseRuName(name string) (ruName string, year string) {
	rx := regexp.MustCompile(` (\([0-9]{4}\))`)

	ruName = name

	if rx.MatchString(name) {
		parts := strings.Split(name, " (")

		if len(parts) >= 2 {
			year = strings.TrimRight(parts[len(parts)-1], ")")
			ruName = strings.Join(parts[:len(parts)-1], " (")
		}
	}

	return ruName, year
}
