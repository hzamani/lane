//go:build latency

package lane

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"testing"
	"time"
)

func TestBytesIndexLatency(t *testing.T) {
	source, err := os.ReadFile("testdata/source")
	if err != nil {
		t.Fatal(err)
	}
	github, err := os.ReadFile("testdata/github")
	if err != nil {
		t.Fatal(err)
	}
	enwik8, err := os.ReadFile("testdata/enwik8")
	if err != nil {
		t.Fatal(err)
	}

	o := t.Output()
	fmt.Fprintf(o, "%-11s %4s %5s %4s", "impl", "data", "base", "size")
	fmt.Fprintf(o, "%8s%8s%8s%8s%8s%8s%8s%8s\n",
		"min",
		"avg",
		"p50",
		"p90",
		"p99",
		"p99.9",
		"max",
		"stddev",
	)
	for base, batchSize := range map[int]int{256: 1000, 4096: 500, 65536: 100, 1044480: 10} {
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
				for i := range needles {
					needles[i] = haystack.data[n : n+size]
					n++
				}

				for _, impl := range []struct {
					name string
					f    func(haystack, needle []byte) int
				}{
					{"bytes.Index", bytes.Index},
					{"BytesIndex", BytesIndex},
				} {
					latencies := make([]time.Duration, len(needles))
					for i := range needles {
						start := time.Now()
						for range batchSize {
							_ = impl.f(haystack.data, needles[i])
						}
						latencies[i] = time.Since(start) / time.Duration(batchSize)
					}
					fmt.Fprintf(o, "%-11s %4s %5d %4d", impl.name, haystack.name, base, size)
					reportLatency(o, latencies)
				}
				latencies := make([]time.Duration, len(needles))
				for i := range needles {
					f := NewBytesFinder(needles[i])
					start := time.Now()
					for range batchSize {
						_ = f.Index(haystack.data)
					}
					latencies[i] = time.Since(start) / time.Duration(batchSize)
				}
				fmt.Fprintf(o, "%-11s %4s %5d %4d", "BytesFinder", haystack.name, base, size)
				reportLatency(o, latencies)
			}
		}
	}
}

func reportLatency(o io.Writer, latencies []time.Duration) {
	count := len(latencies)

	slices.Sort(latencies)

	var sum time.Duration
	for _, d := range latencies {
		sum += d
	}

	mean := float64(sum) / float64(count)
	var sumSq float64
	for _, latency := range latencies {
		d := float64(latency) - mean
		sumSq += d * d
	}
	stdDev := time.Duration(math.Sqrt(sumSq / float64(count)))

	fmt.Fprintf(o, "%8d%8d%8d%8d%8d%8d%8d%8d\n",
		latencies[0].Nanoseconds(),
		int64(mean),
		latencies[len(latencies)*50/100].Nanoseconds(),
		latencies[len(latencies)*90/100].Nanoseconds(),
		latencies[len(latencies)*99/100].Nanoseconds(),
		latencies[len(latencies)*999/1000].Nanoseconds(),
		latencies[len(latencies)-1].Nanoseconds(),
		stdDev.Nanoseconds(),
	)
}
