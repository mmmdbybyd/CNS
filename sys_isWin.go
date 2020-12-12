// +build windows

// isWin.go
package main

import (
	"log"
	"net"
	"syscall"
)

func setMaxNofile() {
}

func setsid() {
}

func enableTcpFastopen(listener net.Listener) {
	const CNS_TCP_FASTOPEN int = 0x17
	f, err := listener.(*net.TCPListener).File()
	if err != nil {
		log.Println(err)
		return
	}
	if err := syscall.SetsockoptInt(syscall.Handle(f.Fd()), syscall.IPPROTO_TCP, CNS_TCP_FASTOPEN, 1); err != nil {
		log.Println(err)
	}
	f.Close()
}
