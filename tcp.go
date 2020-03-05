// tcp.go
package main

import (
	"bytes"
	//"fmt"
	"log"
	"net"
	"time"
)

/* 把fromConn的数据转发到toConn */
func tcpForward(fromConn, toConn net.Conn, payload []byte, ENorDE_Func func([]byte, byte) byte, dynamic_code byte) {
	defer fromConn.Close()
	defer toConn.Close()

	var RLen, WLen int
	for {
		fromConn.SetReadDeadline(time.Now().Add(tcp_timeout))
		if RLen, _ = fromConn.Read(payload[:]); RLen <= 0 {
			return
		}
		if passLen_code != 0 {
			dynamic_code = ENorDE_Func(payload[:RLen], dynamic_code)
		}
		toConn.SetWriteDeadline(time.Now().Add(tcp_timeout))
		if WLen, _ = toConn.Write(payload[:RLen]); WLen <= 0 {
			return
		}
	}
}

/* 从header中获取host */
func getProxyHost(header []byte) (string, byte) {
	hostSub := bytes.Index(header, proxyKey)
	if hostSub < 0 {
		return "", 0
	}
	hostSub += len(proxyKey)
	hostEndSub := bytes.IndexByte(header[hostSub:], '\r')
	if hostEndSub < 0 {
		return "", 0
	}
	hostEndSub += hostSub
	if passLen_code != 0 {
		host, dynamic_code, err := CuteBi_decrypt_host(header[hostSub:hostEndSub])
		if err != nil {
			log.Println(err)
			return "", 0
		}
		return string(host), dynamic_code
	} else {
		return string(header[hostSub:hostEndSub]), 0
	}
}

/* 处理tcp会话 */
func handleTcpSession(cConn *net.TCPConn, header []byte) {
	defer log.Println("A tcp client close")

	/* 获取请求头中的host */
	host, dynamic_code := getProxyHost(header)
	if host == "" {
		log.Println("No proxy host: {" + string(header) + "}")
		cConn.Write([]byte("No proxy host"))
		cConn.Close()
		return
	}
	log.Println("proxyHost: " + host)
	/* 连接目标地址 */
	sAddr, _ := net.ResolveTCPAddr("tcp", host)
	sConn, err := net.DialTCP("tcp", nil, sAddr)
	if err != nil {
		log.Println(err)
		cConn.Write([]byte("Proxy address [" + host + "] error"))
		cConn.Close()
		return
	}
	sConn.SetKeepAlive(true)
	sConn.SetKeepAlivePeriod(tcp_keepAlive)
	/* 开始转发 */
	log.Println("Start tcpForward")
	go tcpForward(cConn, sConn, make([]byte, 8192), CuteBi_decrypt, dynamic_code)
	tcpForward(sConn, cConn, header, CuteBi_encrypt, dynamic_code)
}
