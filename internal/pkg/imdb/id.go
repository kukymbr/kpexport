package imdb

type TitleID string

func (id *TitleID) String() string {
	return string(*id)
}
