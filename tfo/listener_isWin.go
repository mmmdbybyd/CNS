// +build windows

/* windows暂时不支持tcpFastOpen */

package tfo

import (
	"net"
)

func Listen(host string) (Listener, error) {
	return net.Listen("tcp", host)
}
