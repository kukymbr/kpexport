package domain

import "fmt"

type Votes []Vote

func (v *Votes) Add(vote Vote) error {
	if vote.MovieURL == "" {
		return fmt.Errorf("no movie URL is vote item %s", vote.MovieNameRu)
	}

	if vote.Rate == 0 {
		return fmt.Errorf("no movie rate is vote item %s", vote.MovieNameRu)
	}

	*v = append(*v, vote)

	return nil
}

func (v *Votes) AddOnce(vote Vote) error {
	for _, curr := range *v {
		if curr.MovieURL == vote.MovieURL {
			return nil
		}
	}

	return v.Add(vote)
}
