package sure

import (
	"syscall"
)

func getSysTimes(sys *syscall.Stat_t) (int64, int64) {
	return sys.Mtimespec.Sec, sys.Ctimespec.Sec
}
