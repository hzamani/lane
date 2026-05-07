package lane

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func addBytesIndexSeeds(f *testing.F) {
	withOne := func(size int, at int) []byte {
		data := make([]byte, size)
		data[at] = 1
		return data
	}

	f.Add([]byte(""), []byte(""))                           // both empty
	f.Add([]byte(""), []byte("x"))                          // empty haystack
	f.Add([]byte("x"), []byte(""))                          // empty needle
	f.Add([]byte("x"), []byte("x"))                         // length 1, exact match
	f.Add([]byte("xy"), []byte("x"))                        // length 1 needle
	f.Add([]byte("xy"), []byte("y"))                        // length 1 needle, end
	f.Add([]byte("aaaaaaaa"), []byte("aaa"))                // repetitive
	f.Add([]byte("aaaaaaaa"), []byte("aab"))                // near-match
	f.Add(bytes.Repeat([]byte("ab"), 50), []byte("ababab")) // periodic
	f.Add(bytes.Repeat([]byte("a"), 100), []byte("a"))      // long, single char
	f.Add(withOne(32, 31), []byte{1})                       // boundary at vector width
	f.Add(withOne(33, 32), []byte{1})                       // just past vector width
	f.Add(withOne(64, 63), []byte{1})                       // two vectors
	f.Add(withOne(65, 64), []byte{1})                       // just past second vector
	f.Add([]byte("this is standard english text"), []byte("standard"))
	f.Add([]byte(`{"key": "value", "status": "ok"}`), []byte(`"status": "ok"`))
	f.Add([]byte("func Index(haystack, needle []byte) int {"), []byte("needle"))
}

func FuzzBytesIndex(f *testing.F) {
	addBytesIndexSeeds(f)

	f.Fuzz(func(t *testing.T, haystack, needle []byte) {
		want := bytes.Index(haystack, needle)
		got := BytesIndex(haystack, needle)
		if want != got {
			t.Errorf("want = %d, got %d\n  haystack(%d)=%x\n  needle(%d)=%x",
				want, got, len(haystack), haystack, len(needle), needle,
			)
		}
	})
}

func FuzzBytesFinderIndex(f *testing.F) {
	addBytesIndexSeeds(f)

	f.Fuzz(func(t *testing.T, haystack, needle []byte) {
		want := bytes.Index(haystack, needle)
		got := NewBytesFinder(needle).Index(haystack)
		if want != got {
			t.Errorf("want = %d, got %d\n  haystack(%d)=%x\n  needle(%d)=%x",
				want, got, len(haystack), haystack, len(needle), needle,
			)
		}
	})
}

func FuzzBytesFinderByIndex(f *testing.F) {
	addBytesIndexSeeds(f)

	rank := ByteRanker{}
	for i := range rank {
		rank[i] = uint8(i / 2)
	}

	f.Fuzz(func(t *testing.T, haystack, needle []byte) {
		want := bytes.Index(haystack, needle)
		got := NewBytesFinderBy(needle, rank).Index(haystack)
		if want != got {
			t.Errorf("want = %d, got %d\n  haystack(%d)=%x\n  needle(%d)=%x",
				want, got, len(haystack), haystack, len(needle), needle,
			)
		}
	})
}

func BenchmarkBytesIndex(b *testing.B) {
	source, err := os.ReadFile("testdata/source")
	if err != nil {
		b.Fatal(err)
	}
	github, err := os.ReadFile("testdata/github")
	if err != nil {
		b.Fatal(err)
	}
	enwik8, err := os.ReadFile("testdata/enwik8")
	if err != nil {
		b.Fatal(err)
	}

	for _, base := range []int{256, 4096, 65536, 1044480} {
		for _, size := range []int{2, 4, 18, 80, 140} {
			for _, haystack := range []struct {
				name string
				data []byte
			}{
				{"json", github},
				{"code", source},
				{"wiki", enwik8},
			} {
				needles := make([][]byte, 1024)
				n := max(0, base-512)
				var indexes int64
				for i := range 1024 / 8 {
					for j := range 8 {
						needles[i+j] = haystack.data[n : n+size+j]
						indexes += int64(BytesIndex(haystack.data, needles[i+j]))
						n++
					}
				}

				for _, impl := range []struct {
					name string
					f    func(haystack, needle []byte) int
				}{
					{"bytes.Index", bytes.Index},
					{"BytesIndex", BytesIndex},
				} {
					name := fmt.Sprintf("impl=%s/data=%s/base=%d/size=%d",
						impl.name, haystack.name, base, size)
					b.Run(name, func(b *testing.B) {
						b.SetBytes(indexes / 1024)
						for i := range b.N {
							_ = impl.f(haystack.data, needles[i&1023])
						}
					})
				}

				finders := make([]*BytesFinder, 1024)
				for i := range needles {
					finders[i] = NewBytesFinder(needles[i])
				}
				name := fmt.Sprintf("impl=BytesFinder/data=%s/base=%d/size=%d", haystack.name, base, size)
				b.Run(name, func(b *testing.B) {
					b.SetBytes(indexes / 1024)
					for i := range b.N {
						_ = finders[i&1023].Index(haystack.data)
					}
				})
			}
		}
	}
}
