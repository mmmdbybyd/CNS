// +build !windows

package main

import (
	"syscall"
)

func setMaxNofile() {
	syscall.Setrlimit(syscall.RLIMIT_NOFILE, &syscall.Rlimit{Cur: 1048576, Max: 1048576})
}

func setsid() {
	syscall.Setsid()
}
