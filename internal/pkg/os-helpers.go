package pkg

import "runtime"

type Platform struct {
	OS   string
	Arch string
}

func HostPlatform() Platform {
	return Platform{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}
