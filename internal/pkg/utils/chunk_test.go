package utils_test

import (
	"testing"

	"github.com/kukymbr/kinopoiskexport/internal/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChunk(t *testing.T) {
	items := []string{"test1", "test2", "test3", "test4", "test5"}

	chunks := utils.Chunk(items, 2)

	require.Len(t, chunks, 3)

	assert.Len(t, chunks[0], 2)
	assert.Len(t, chunks[1], 2)
	assert.Len(t, chunks[2], 1)

	assert.Equal(t, "test1", chunks[0][0])
	assert.Equal(t, "test3", chunks[1][0])
	assert.Equal(t, "test5", chunks[2][0])
}

func TestChunk_WhenChunkSizeGreaterThenSize(t *testing.T) {
	items := []string{"test1", "test2", "test3", "test4", "test5"}

	chunks := utils.Chunk(items, 6)

	require.Len(t, chunks, 1)
	assert.Len(t, chunks[0], 5)
	assert.Equal(t, "test1", chunks[0][0])
	assert.Equal(t, "test5", chunks[0][4])
}
