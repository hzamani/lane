//go:build !goexperiment.simd && !amd64.v3 && !amd64.v4

package lane

var (
	hasAVX2   = false
	hasAVX512 = false
)
