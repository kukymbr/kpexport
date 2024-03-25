//go:build unix

package utils_test

import (
	"testing"

	"github.com/kukymbr/kinopoiskexport/internal/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestFixSeparators(t *testing.T) {
	tests := []struct {
		Input    string
		Expected string
	}{
		{"", ""},
		{`/home\user`, `/home/user`},
		{`/usr\local/../local\/bin`, `/usr/local/../local//bin`},
	}

	for i, test := range tests {
		path := utils.FixSeparators(test.Input)

		assert.Equal(t, test.Expected, path, i)
	}
}
