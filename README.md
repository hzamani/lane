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
cpu: AMD Ryzen 7 4800HSS with Radeon Graphics
        │   bytes.Index    │                BytesIndex                │               BytesFinder                │
        │      sec/op      │     sec/op       vs base                 │     sec/op       vs base                 │
json      278.55n ±  43% ¹   48.58n ±  29% ¹  -82.56% (p=0.000 n=120)   38.62n ±  43% ¹  -86.14% (p=0.000 n=120)
code      335.20n ± 418% ¹   64.48n ± 501% ¹  -80.76% (p=0.000 n=120)   55.16n ± 512% ¹  -83.55% (p=0.000 n=120)
wiki      340.75n ± 105% ¹   52.48n ± 171% ¹  -84.60% (p=0.000 n=120)   41.24n ± 204% ¹  -87.90% (p=0.000 n=120)
geomean    316.9n            54.78n           -82.71%                   44.45n           -85.97%
¹ benchmarks vary in .fullname

        │   bytes.Index   │                BytesIndex                 │                BytesFinder                 │
        │       B/s       │       B/s        vs base                  │       B/s         vs base                  │
json      18.30Gi ±  6% ¹   78.51Gi ±  5% ¹  +328.93% (p=0.000 n=120)   111.62Gi ± 12% ¹  +509.84% (p=0.000 n=120)
code      16.02Gi ± 11% ¹   74.38Gi ± 11% ¹  +364.36% (p=0.000 n=120)   105.70Gi ±  4% ¹  +559.95% (p=0.000 n=120)
wiki      15.64Gi ± 17% ¹   73.90Gi ±  3% ¹  +372.59% (p=0.000 n=120)   103.03Gi ± 10% ¹  +558.88% (p=0.000 n=120)
geomean   16.61Gi           75.57Gi          +354.89%                    106.7Gi          +542.46%
¹ benchmarks vary in .fullname
```

Execution variance of finding a 140 bytes needle in 1MB of go source code.

```txt
              min    avg    p50     p90     p99   p99.9     max   stddev
bytes.Index  35µs  492µs  334µs  1256µs  1261µs  1264µs  1264µs  444.8µs
BytesIndex   70µs   83µs   76µs    96µs   187µs   191µs   191µs   19.8µs
BytesFinder  64µs   65µs   65µs    66µs    67µs    68µs    68µs    0.8µs
```
