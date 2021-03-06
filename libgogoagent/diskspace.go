package libgogoagent

import (
	"strconv"
	"syscall"
)

func UsableSpace() string {
	_, free, err := diskSpace("/")
	if err != nil {
		LogInfo("Unknown diskspace, error: %v", err)
		return "-1"
	}
	return strconv.Itoa(free)
}

// Space returns total and free bytes available in a directory, e.g. `/`.
// Think of it as "df" UNIX command.
func diskSpace(path string) (total, free int, err error) {
	s := syscall.Statfs_t{}
	err = syscall.Statfs(path, &s)
	if err != nil {
		return
	}
	total = int(s.Bsize) * int(s.Blocks)
	free = int(s.Bsize) * int(s.Bfree)
	return
}
