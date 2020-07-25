// +build !windows

package tfo

import (
	"bytes"
	"net"
	"os"
	"syscall"
)

const (
	TCPFastOpen   int = 23
	ListenBacklog int = 23
)

// tfoListener implement the RazorListener can also be used as the net.Listener
type tfoListener struct {
	ServerAddr [16]byte
	ServerPort int
	fd         int
}

// Listen will listen given host and give back a Listener which implement the net.Listener
func Listen(host string) (Listener, error) {
	r := &tfoListener{}

	addr, err := net.ResolveTCPAddr("tcp6", host)
	if err == nil {
		//addr.IP存放的是ipv6地址，直接复制
		copy(r.ServerAddr[:], addr.IP)
	} else {
		//不是ipv6地址，尝试解析ipv4地址
		addr, err = net.ResolveTCPAddr("tcp4", host)
		if err != nil {
			return nil, err
		}
		if bytes.HasSuffix(addr.IP, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) {
			//addr.IP前4字节存放ipv4地址，转为ipv4映射ipv6
			copy(r.ServerAddr[:12], []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff, 0xff})
			copy(r.ServerAddr[12:], addr.IP[:4])
		} else {
			//addr.IP存放的是ipv4映射ipv6地址  直接复制
			copy(r.ServerAddr[:], addr.IP)
		}
	}
	r.ServerPort = addr.Port

	r.fd, err = syscall.Socket(syscall.AF_INET6, syscall.SOCK_STREAM, 0)
	if err != nil {
		if err == syscall.ENOPROTOOPT {
			return nil, ErrTFONotSupport
		}
		return nil, err
	}
	syscall.SetsockoptInt(r.fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	err = syscall.Bind(r.fd, &syscall.SockaddrInet6{Addr: r.ServerAddr, Port: r.ServerPort})
	if err != nil {
		return nil, err
	}
	err = syscall.SetsockoptInt(r.fd, syscall.IPPROTO_TCP, TCPFastOpen, 3)
	if err != nil {
		return nil, err
	}
	err = syscall.Listen(r.fd, ListenBacklog)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (r *tfoListener) Accept() (net.Conn, error) {
	cfd, _, err := syscall.Accept(r.fd)
	if err != nil {
		return nil, err
	}

	f := os.NewFile(uintptr(cfd), "")
	defer f.Close()
	return net.FileConn(f)
}

func (r *tfoListener) Close() error {
	err := syscall.Shutdown(r.fd, syscall.SHUT_RDWR)
	if err != nil {
		return err
	}
	err = syscall.Close(r.fd)
	if err != nil {
		return err
	}
	return nil
}

func (r *tfoListener) Addr() net.Addr {
	return &net.TCPAddr{
		IP:   r.ServerAddr[:],
		Port: r.ServerPort,
	}
}
