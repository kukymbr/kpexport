package utils

// Chunk splits slice into the chunks with defined chunkSize.
func Chunk[T any](items []T, chunkSize uint) [][]T {
	size := int(chunkSize)
	chunks := make([][]T, 0, (len(items)+size-1)/size)

	for size < len(items) {
		items, chunks = items[size:], append(chunks, items[0:size:size])
	}

	return append(chunks, items)
}
