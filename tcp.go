package main

import (
	"bytes"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

var tcpBufferPool sync.Pool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 8192)
	},
}

/* 把fromConn的数据转发到toConn */
func tcpForward(fromConn, toConn net.Conn, payload []byte) {
	defer func() {
		fromConn.Close()
		toConn.Close()
	}()

	var RLen, WLen, CuteBi_XorCrypt_passwordSub int
	var err error
	for {
		fromConn.SetReadDeadline(time.Now().Add(config.Tcp_timeout))
		toConn.SetReadDeadline(time.Now().Add(config.Tcp_timeout))
		if RLen, err = fromConn.Read(payload); err != nil || RLen <= 0 {
			return
		}
		if len(CuteBi_XorCrypt_password) != 0 {
			CuteBi_XorCrypt_passwordSub = CuteBi_XorCrypt(payload[:RLen], CuteBi_XorCrypt_passwordSub)
		}
		toConn.SetWriteDeadline(time.Now().Add(config.Tcp_timeout))
		if WLen, err = toConn.Write(payload[:RLen]); err != nil || WLen <= 0 {
			return
		}
	}
}

/* 从header中获取host */
func getProxyHost(header []byte) string {
	hostSub := bytes.Index(header, []byte(config.Proxy_key))
	if hostSub < 0 {
		return ""
	}
	hostSub += len(config.Proxy_key)
	hostEndSub := bytes.IndexByte(header[hostSub:], '\r')
	if hostEndSub < 0 {
		return ""
	}
	hostEndSub += hostSub
	if len(CuteBi_XorCrypt_password) != 0 {
		host, err := CuteBi_decrypt_host(header[hostSub:hostEndSub])
		if err != nil {
			log.Println(err)
			return ""
		}
		return string(host)
	} else {
		return string(header[hostSub:hostEndSub])
	}
}

/* 处理tcp会话 */
func handleTcpSession(cConn net.Conn, header []byte) {
	// defer log.Println("A tcp client close")

	/* 获取请求头中的host */
	host := getProxyHost(header)
	if host == "" {
		log.Println("No proxy host: {" + string(header) + "}")
		cConn.Write([]byte("No proxy host"))
		cConn.Close()
		return
	}
	// log.Println("proxyHost: " + host)
	//tcpDNS over udpDNS
	if config.Enable_dns_tcpOverUdp && strings.HasSuffix(host, ":53") == true {
		dns_tcpOverUdp(cConn, host, header)
		return
	}
	/* 连接目标地址 */
	if strings.Contains(host, ":") == false {
		host += ":80"
	}
	sConn, dialErr := net.Dial("tcp", host)
	if dialErr != nil {
		log.Println(dialErr)
		cConn.Write([]byte("Proxy address [" + host + "] DialTCP() error"))
		cConn.Close()
		return
	}
	/* 开始转发 */
	// log.Println("Start tcpForward")

	go tcpForward(cConn, sConn, header)
	newBuff := tcpBufferPool.Get().([]byte)
	tcpForward(sConn, cConn, newBuff)
	tcpBufferPool.Put(newBuff)
}
