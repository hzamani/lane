# Lane

Lane is an experimental package built to explore SIMD algorithms and measure
their effectiveness in Go.

Powered by the experimental `simd/archsimd` package introduced in Go 1.26.

## Goals

1. Keep it clean and readable.
2. Measure effectiveness.
3. Zero Assembly.

## Performance Summary

`lane` is benchmarked against Go's standard library using large,
real-world datasets (JSON, Go source code, and Wikipedia dumps).
Full benchmark outputs and latency distributions can be found in the
`benchmarks/` directory.

## Implemented Functions

While a few functions in the standard Go library (such as `bytes.Equal` or
`bytes.IndexByte`) utilize SIMD via hand-written assembly, most complex
functions do not—likely due to the high maintenance cost. `lane` tries to
implement those that are not utilizing SIMD.

### Byte slices

* `BytesIndex(haystack, needle []byte) int`

Uses a vectorized sub-slice search. It uses the first and the last byte as a
SIMD stencil to rapidly skip over non-matching sections of the haystack, doing
equality check only when the exact first and last bytes are found.

* `NewBytesFinder(needle []byte) *BytesFinder`
* `NewBytesFinderBy(needle []byte, ranker ByteRanker) *BytesFinder`

Builds a reusable searcher for a fixed needle. Precomputes a rare-pair stencil
using the byte ranker to reduce false positives — more effective than `BytesIndex`
for repeated searches on the same needle. `ByteRankerDefault` is calibrated for
English text and source code; supply a custom `ByteRanker` for other input distributions.

```txt
goos: linux
goarch: amd64
pkg: github.com/hzamani/lane
cpu: AMD Ryzen AI 7 PRO 360 w/ Radeon 880M
        │   bytes.Index    │                BytesIndex                │               BytesFinder                │
        │      sec/op      │     sec/op       vs base                 │     sec/op       vs base                 │
json      158.60n ±  59% ¹   15.04n ±  45% ¹  -90.52% (p=0.000 n=120)   13.60n ±  43% ¹  -91.42% (p=0.000 n=120)
code      242.90n ± 357% ¹   21.81n ± 458% ¹  -91.02% (p=0.000 n=120)   18.98n ± 462% ¹  -92.19% (p=0.000 n=120)
wiki      250.40n ±  67% ¹   19.28n ± 144% ¹  -92.30% (p=0.000 n=120)   13.94n ± 223% ¹  -94.43% (p=0.000 n=120)
geomean    212.9n            18.49n           -91.31%                   15.32n           -92.80%
¹ benchmarks vary in .fullname

        │   bytes.Index   │                 BytesIndex                 │                 BytesFinder                 │
        │       B/s       │       B/s         vs base                  │       B/s         vs base                   │
json      33.50Gi ± 24% ¹   239.94Gi ± 14% ¹  +616.29% (p=0.000 n=120)   355.06Gi ± 20% ¹   +959.96% (p=0.000 n=120)
code      28.93Gi ± 18% ¹   210.04Gi ±  6% ¹  +626.15% (p=0.000 n=120)   327.57Gi ± 11% ¹  +1032.44% (p=0.000 n=120)
wiki      27.40Gi ± 21% ¹   206.47Gi ± 10% ¹  +653.42% (p=0.000 n=120)   313.28Gi ± 14% ¹  +1043.18% (p=0.000 n=120)
geomean   29.83Gi            218.3Gi          +631.79%                    331.5Gi          +1011.24%
¹ benchmarks vary in .fullname
```

Execution variance of finding a 140 bytes needle in 1MB of go source code.

```txt
              min    avg    p50    p90    p99  p99.9    max   stddev
bytes.Index  19µs  381µs  192µs 1041µs 1598µs 1667µs 1768µs  433.3µs
BytesIndex   19µs   32µs   29µs   47µs   89µs  133µs  171µs   14.0µs
BytesFinder  18µs   19µs   18µs   19µs   35µs   61µs   69µs    3.8µs
```
