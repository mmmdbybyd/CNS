// +build !windows

// isNotWin.go
package main

import (
	"fmt"
	/*
		"log"
		"net"
	*/
	"syscall"
)

func setMaxNofile() {
	var rlim syscall.Rlimit
	rlim.Cur = 1048576
	rlim.Max = 1048576
	err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rlim)
	if err != nil {
		fmt.Println(err)
	}
}

func setsid() {
	syscall.Setsid()
}

/*
func enableTcpFastopen(listener *net.TCPListener) {
	const CNS_TCP_FASTOPEN int = 0x17
	f, _ := listener.File()
	if err := syscall.SetsockoptInt(int(f.Fd()), syscall.IPPROTO_TCP, CNS_TCP_FASTOPEN, 1); err != nil {
		log.Println(err)
	}
}
*/
