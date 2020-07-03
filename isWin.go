// +build windows

// isWin.go
package main

/*
	"log"
	"net"
	"syscall"
*/

func setMaxNofile() {
}

func setsid() {
}

/*
func enableTcpFastopen(listener *net.TCPListener) {
	const CNS_TCP_FASTOPEN int = 0x17
	f, _ := listener.File()
	if err := syscall.SetsockoptInt(syscall.Handle(f.Fd()), syscall.IPPROTO_TCP, CNS_TCP_FASTOPEN, 1); err != nil {
		log.Println(err)
	}
}
*/
