//go:build !goexperiment.simd

package lane

import "bytes"

// NewBytesFinderBy falls back to bytes.Index on non-SIMD builds; ranker is unused.
func NewBytesFinderBy(needle []byte, ranker ByteRanker) *BytesFinder {
	return &BytesFinder{
		needle: needle,
		index: func(haystack []byte) int {
			return bytes.Index(haystack, needle)
		},
	}
}

// BytesIndex falls back to bytes.Index on non-SIMD builds.
func BytesIndex(haystack, needle []byte) int {
	return bytes.Index(haystack, needle)
}
