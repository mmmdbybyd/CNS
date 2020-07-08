/* By: %73 */

package tfo

import (
	"errors"
	"net"
)

var (
	// ErrTFONotSupport give out an error your OS kernel is not support the tfo
	ErrTFONotSupport = errors.New("TCP Fast Open server support is unavailable (unsupported kernel)")
	// ErrParseHost mean can not parse given host into ip:port
	ErrParseHost = errors.New("Error parse host")
)

// Listener accept tcp connections from binding socket implements net.Listener
type Listener interface {
	net.Listener
}

// Conn is a tcp fast open supported Conn which implement net.Conn
type Conn interface {
	net.Conn
}
