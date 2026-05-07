//go:build amd64 && !amd64.v3

package lane

import "golang.org/x/sys/cpu"

var (
	hasAVX2 = cpu.X86.HasAVX2
)
