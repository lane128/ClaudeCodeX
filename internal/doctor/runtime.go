package doctor

import "runtime"

func runtimeGOOS() string {
	return runtime.GOOS
}
