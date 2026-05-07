package lane

import (
	"encoding/binary"
	"math/rand/v2"
	"testing"
)

func randBytes(n int) []byte {
	b := make([]byte, n)
	i := 0
	for i+8 < n {
		binary.LittleEndian.PutUint64(b[i:], rand.Uint64())
		i += 8
	}
	for i+4 < n {
		binary.LittleEndian.PutUint32(b[i:], rand.Uint32())
		i += 4
	}
	for i < n {
		b[i] = byte(rand.Uint32() & 0xFF)
		i++
	}
	return b
}

func BenchmarkByteRank(b *testing.B) {
	needles := make([][]byte, 1024)
	for i := range needles {
		needles[i] = randBytes(rand.IntN(1024))
	}

	b.ResetTimer()
	for i := range b.N {
		_, _ = ByteRankerDefault.pickTwo(needles[i&1023])
	}
}

func FuzzByteRank(f *testing.F) {
	rank := ByteRanker{}
	for i := range rank {
		rank[i] = uint8(i / 2)
	}

	f.Add([]byte{1, 2})
	f.Add([]byte{1, 1})
	f.Add([]byte{2, 1})
	f.Add([]byte{1, 1, 1, 1})
	f.Fuzz(func(t *testing.T, needle []byte) {
		if len(needle) < 2 {
			return
		}

		x, y := rank.pickTwo(needle)
		if x >= y {
			t.Errorf("pickTwo(%x) returned (%d, %d)", needle, x, y)
		}
	})
}

func FuzzPickTwoOptimality(f *testing.F) {
	rank := ByteRanker{}
	for i := range rank {
		rank[i] = uint8(i / 2)
	}

	f.Add([]byte{1, 2})
	f.Add([]byte{1, 1})
	f.Add([]byte{2, 1})
	f.Add([]byte{1, 1, 1, 1})
	f.Fuzz(func(t *testing.T, needle []byte) {
		if len(needle) < 2 {
			return
		}

		x, y := rank.pickTwo(needle)
		selected := uint16(rank[needle[x]]) + uint16(rank[needle[y]])

		for i := range len(needle) {
			for j := i + 1; j < len(needle); j++ {
				if combined := uint16(rank[needle[i]]) + uint16(rank[needle[j]]); combined < selected {
					t.Errorf("pickTwo(%x) = (%d, %d) combined rank %d, but pair (%d, %d) has rank %d",
						needle, x, y, selected, i, j, combined)
					return
				}
			}
		}
	})
}
