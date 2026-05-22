//go:build goexperiment.simd && !amd64.v3 && !amd64.v4

package lane

import "simd/archsimd"

var (
	hasAVX2   = archsimd.X86.AVX2()
	hasAVX512 = archsimd.X86.AVX512()
)
