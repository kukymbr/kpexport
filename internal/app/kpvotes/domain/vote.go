package domain

import "time"

type Vote struct {
	MovieURL          string
	MovieNameRu       string
	MovieNameOriginal string
	MovieYear         string

	Timestamp time.Time

	Rate uint8

	ImdbID string
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
