package lane

// BytesFinder precomputes a rare-pair stencil for repeated searches against
// a fixed needle.
type BytesFinder struct {
	needle []byte
	x, y   int
	index  func([]byte) int
}

// NewBytesFinder builds a BytesFinder using ByteRankerDefault to select
// the stencil pair.
func NewBytesFinder(needle []byte) *BytesFinder {
	return NewBytesFinderBy(needle, ByteRankerDefault)
}

// Index returns the index of the first instance of needle in haystack,
// or -1 if it is not present.
func (f *BytesFinder) Index(haystack []byte) int {
	return f.index(haystack)
}
