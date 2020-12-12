// +build !windows

package main

import (
	"log"
	"net"
	"syscall"
)

func setMaxNofile() {
	syscall.Setrlimit(syscall.RLIMIT_NOFILE, &syscall.Rlimit{Cur: 1048576, Max: 1048576})
}

func setsid() {
	syscall.Setsid()
}

func enableTcpFastopen(listener net.Listener) {
	const CNS_TCP_FASTOPEN int = 0x17
	f, _ := listener.(*net.TCPListener).File()
	if err := syscall.SetsockoptInt(int(f.Fd()), syscall.IPPROTO_TCP, CNS_TCP_FASTOPEN, 1); err != nil {
		log.Println(err)
	}
	f.Close()
}
