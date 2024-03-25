package domain

import (
	"time"

	"github.com/kukymbr/kinopoiskexport/internal/pkg/imdb"
)

type Vote struct {
	MovieURL          string
	MovieNameRu       string
	MovieNameOriginal string
	MovieYear         string

	Timestamp time.Time

	Rate uint8

	ImdbID imdb.TitleID
}

func (v *Vote) GetOriginalTitle() string {
	title := v.MovieNameRu

	if v.MovieNameOriginal != "" {
		title = v.MovieNameOriginal
	}

	if v.MovieYear != "" {
		title += " (" + v.MovieYear + ")"
	}

	return title
}

func (v *Vote) GetIMDbURL() string {
	if v.ImdbID == "" {
		return ""
	}

	return imdb.TitleURL + v.ImdbID.String()
}
