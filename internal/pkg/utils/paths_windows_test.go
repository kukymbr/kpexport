//go:build windows

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
		{`C:/data/`, `C:\data\`},
		{`C:/data\test.txt`, `C:\data\test.txt`},
	}

	for i, test := range tests {
		path := utils.FixSeparators(test.Input)

		assert.Equal(t, test.Expected, path, i)
	}
}
