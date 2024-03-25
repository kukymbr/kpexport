package imdb

import "regexp"

var titleIDRx = regexp.MustCompile(`^tt([0-9]+)$`)

type TitleID string

func (id *TitleID) IsValid() bool {
	return titleIDRx.MatchString(id.String())
}

func (id *TitleID) String() string {
	return string(*id)
}

func (id *TitleID) ToURL() string {
	return Host + "/title/" + id.String()
}
