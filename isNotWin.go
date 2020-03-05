// +build !windows

// isNotWin.go
package main

import (
	"fmt"
	"log"
	"net"
	"syscall"
)

func setMaxNofile() {
	return
	var rlim syscall.Rlimit
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rlim)
	if err != nil {
		fmt.Println(err)
	}
	rlim.Cur = 65535
	rlim.Max = 65535
	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlim)
	if err != nil {
		fmt.Println(err)
	}
}

func setsid() {
	syscall.Setsid()
}

func enableTcpFastopen(listener *net.TCPListener) {
	const CNS_TCP_FASTOPEN int = 0x17
	f, _ := listener.File()
	if err := syscall.SetsockoptInt(int(f.Fd()), syscall.IPPROTO_TCP, CNS_TCP_FASTOPEN, 1); err != nil {
		log.Println(err)
	}
}
