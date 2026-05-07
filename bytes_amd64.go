//go:build amd64 && goexperiment.simd

package lane

import (
	"bytes"
	"math/bits"
	"simd/archsimd"
)

// NewBytesFinderBy lets callers supply a domain-specific ranker for better
// stencil pair selection.
func NewBytesFinderBy(needle []byte, ranker ByteRanker) *BytesFinder {
	f := &BytesFinder{needle: needle}
	if !hasAVX2 {
		f.index = func(haystack []byte) int {
			return bytes.Index(haystack, f.needle)
		}
		return f
	}
	switch len(needle) {
	case 0:
		f.index = func([]byte) int {
			return 0
		}
		return f
	case 1:
		f.index = func(haystack []byte) int {
			return bytes.IndexByte(haystack, needle[0])
		}
		return f
	case 2:
		f.y = 1
	default:
		f.x, f.y = ranker.pickTwo(needle)
	}
	f.index = func(haystack []byte) int {
		n := len(f.needle)
		h := len(haystack)
		if n >= h {
			if bytes.Equal(haystack, f.needle) {
				return 0
			}
			return -1
		}
		x, y := f.x, f.y
		nx, ny := f.needle[x], f.needle[y]
		vx := archsimd.BroadcastUint8x32(nx)
		vy := archsimd.BroadcastUint8x32(ny)
		i := 0
		for i <= h-n-32 {
			va := archsimd.LoadUint8x32Slice(haystack[i+x:])
			vb := archsimd.LoadUint8x32Slice(haystack[i+y:])
			mask := va.Equal(vx).And(vb.Equal(vy)).ToBits()
			for mask != 0 {
				pos := i + bits.TrailingZeros32(mask)
				if bytes.Equal(haystack[pos:pos+n], f.needle) {
					return pos
				}
				mask &= mask - 1
			}
			i += 32
		}
		for i <= h-n {
			if haystack[i+x] == nx &&
				haystack[i+y] == ny &&
				bytes.Equal(haystack[i:i+n], f.needle) {
				return i
			}
			i++
		}
		return -1
	}
	return f
}

// BytesIndex returns the index of the first instance of needle in haystack,
// or -1 if it is not present.
// Prefer NewBytesFinder for repeated searches on the same needle.
func BytesIndex(haystack, needle []byte) int {
	if !hasAVX2 {
		return bytes.Index(haystack, needle)
	}
	h := len(haystack)
	n := len(needle)
	x, y := 0, 0
	switch {
	case n >= h:
		if bytes.Equal(haystack, needle) {
			return 0
		}
		return -1
	case n == 0:
		return 0
	case n == 1:
		return bytes.IndexByte(haystack, needle[0])
	default:
		y = n - 1
	}
	vx := archsimd.BroadcastUint8x32(needle[x])
	vy := archsimd.BroadcastUint8x32(needle[y])
	i := 0
	for i <= h-n-32 {
		va := archsimd.LoadUint8x32Slice(haystack[i+x:])
		vb := archsimd.LoadUint8x32Slice(haystack[i+y:])
		mask := va.Equal(vx).And(vb.Equal(vy)).ToBits()
		for mask != 0 {
			pos := i + bits.TrailingZeros32(mask)
			if bytes.Equal(haystack[pos+1:pos+n-1], needle[1:n-1]) {
				return pos
			}
			mask &= mask - 1
		}
		i += 32
	}
	for i <= h-n {
		if haystack[i+x] == needle[x] &&
			haystack[i+y] == needle[y] &&
			bytes.Equal(haystack[i+1:i+n-1], needle[1:n-1]) {
			return i
		}
		i++
	}
	return -1
}
