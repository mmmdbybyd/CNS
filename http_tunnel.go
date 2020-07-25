package main

import (
	"bytes"
	"crypto/tls"
	"log"
	"net"
	"time"

	"./tfo"
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

func handleTunnel(cConn net.Conn, payload []byte, tlsConfig *tls.Config) {
	cConn.SetReadDeadline(time.Now().Add(config.Tcp_timeout))
	RLen, err := cConn.Read(payload)
	if err != nil || RLen <= 0 {
		cConn.Close()
		return
	}
	if isHttpHeader(payload[:RLen]) == false {
		/* 转为tls的conn */
		if tlsConfig != nil {
			cConn = tls.Server(cConn, tlsConfig)
		}
		handleUdpSession(cConn, payload[:RLen])
	} else {
		if config.Enable_httpDNS == false || Respond_HttpDNS(cConn, payload[:RLen]) == false { /*优先处理httpDNS请求*/
			if WLen, err := cConn.Write(rspHeader(payload[:RLen])); err != nil || WLen <= 0 {
				cConn.Close()
				return
			}
			/* 转为tls的conn */
			if tlsConfig != nil {
				cConn = tls.Server(cConn, tlsConfig)
			}
			if bytes.Contains(payload[:RLen], []byte(config.Udp_flag)) == true {
				handleUdpSession(cConn, nil)
			} else {
				handleTcpSession(cConn, payload)
			}
		}
	}
}

func startHttpTunnel(listen_addr string) {
	var (
		listener net.Listener
		conn     net.Conn
		err      error
	)

	if config.Enable_TFO {
		listener, err = tfo.Listen(listen_addr)
	} else {
		listener, err = net.Listen("tcp", listen_addr)
	}
	if err != nil {
		log.Println(err)
		return
	}

	defer listener.Close()

	for {
		conn, err = listener.Accept()
		if err != nil {
			log.Println(err)
			time.Sleep(3 * time.Second)
			continue
		}
		go handleTunnel(conn, make([]byte, 8192), config.Tls.tlsConfig)
	}
}
