package util

import (
	"runtime"
)

var (
	Version   = "dev"
	GitCommit = "n/a"
	BuildDate = ""
	GoVersion = runtime.Version()
	Author    = "Ronnie Flathers @ropnop | Fork by @chin-tech"
)
