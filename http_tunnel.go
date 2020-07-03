// http_tunnel.go
package main

import (
	"bytes"
	"log"
	"net"
	"time"
)

func isHttpHeader(header []byte) bool {
	if bytes.HasPrefix(header, []byte("CONNECT")) == true ||
		bytes.HasPrefix(header, []byte("GET")) == true ||
		bytes.HasPrefix(header, []byte("POST")) == true ||
		bytes.HasPrefix(header, []byte("HEAD")) == true ||
		bytes.HasPrefix(header, []byte("PUT")) == true ||
		bytes.HasPrefix(header, []byte("COPY")) == true ||
		bytes.HasPrefix(header, []byte("DELETE")) == true ||
		bytes.HasPrefix(header, []byte("MOVE")) == true ||
		bytes.HasPrefix(header, []byte("OPTIONS")) == true ||
		bytes.HasPrefix(header, []byte("LINK")) == true ||
		bytes.HasPrefix(header, []byte("UNLINK")) == true ||
		bytes.HasPrefix(header, []byte("TRACE")) == true ||
		bytes.HasPrefix(header, []byte("PATCH")) == true ||
		bytes.HasPrefix(header, []byte("WRAPPED")) == true {
		return true
	}
	return false
}

func rspHeader(header []byte) []byte {
	if bytes.Contains(header, []byte("WebSocket")) == true {
		return []byte("HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: CuteBi Network Tunnel, (%>w<%)\r\n\r\n")
	} else if bytes.HasPrefix(header, []byte("CON")) == true {
		return []byte("HTTP/1.1 200 Connection established\r\nServer: CuteBi Network Tunnel, (%>w<%)\r\nConnection: keep-alive\r\n\r\n")
	} else {
		return []byte("HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\nServer: CuteBi Network Tunnel, (%>w<%)\r\nConnection: keep-alive\r\n\r\n")
	}
}

func handleConn(cConn net.Conn, payload []byte) {
	cConn.SetReadDeadline(time.Now().Add(config.Tcp_timeout))
	RLen, err := cConn.Read(payload)
	if err != nil || RLen <= 0 {
		cConn.Close()
		return
	}
	if isHttpHeader(payload[:RLen]) == false {
		handleUdpSession(cConn, payload[:RLen])
	} else {
		if config.Enable_httpDNS == false || Respond_HttpDNS(cConn, payload[:RLen]) == false { /*优先处理httpDNS请求*/
			if WLen, err := cConn.Write(rspHeader(payload[:RLen])); err != nil || WLen <= 0 {
				cConn.Close()
				return
			}
			if bytes.Contains(payload[:RLen], []byte(config.Udp_flag)) == true {
				handleConn(cConn, payload) //httpUDP需要读取到二进制数据才进行处理
			} else {
				handleTcpSession(cConn, payload)
			}
		}
	}
}

func startHttpTunnel(listen_addr string) {
	listener, err := net.Listen("tcp", listen_addr)
	if err != nil {
		log.Println(err)
		return
	}

	defer listener.Close()

	var conn net.Conn
	for {
		conn, err = listener.Accept()
		if err != nil {
			log.Println(err)
			time.Sleep(3 * time.Second)
			continue
		}
		go handleConn(conn, make([]byte, 8192))
	}
}
