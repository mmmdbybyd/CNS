package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

func dns_tcpOverUdp(cConn net.Conn, host string, buffer []byte) {
	//log.Println("Start dns_tcpOverUdp")
	defer cConn.Close()

	var err error
	var WLen, RLen, payloadLen, CuteBi_XorCrypt_passwordSub int
	var pkgLen uint16
	for {
		cConn.SetReadDeadline(time.Now().Add(config.Tcp_timeout))
		RLen, err = cConn.Read(buffer[payloadLen:])
		if RLen <= 0 || err != nil {
			return
		}
		//解密
		if len(CuteBi_XorCrypt_password) != 0 {
			CuteBi_XorCrypt_passwordSub = CuteBi_XorCrypt(buffer[payloadLen:payloadLen+RLen], CuteBi_XorCrypt_passwordSub)
		}
		payloadLen += RLen
		if payloadLen > 2 {
			pkgLen = (uint16(buffer[0]) << 8) | (uint16(buffer[1])) //包长度转换
			//防止访问非法数据
			if int(pkgLen)+2 > len(buffer) {
				return
			}
			//如果读取到了一个完整的包，就跳出循环
			if int(pkgLen)+2 <= payloadLen {
				break
			}
		}
	}
	/* 连接目标地址 */
	sConn, dialErr := net.Dial("udp", host)
	if dialErr != nil {
		log.Println(dialErr)
		cConn.Write([]byte("Proxy address [" + host + "] DNS Dial() error"))
		return
	}
	defer sConn.Close()
	if WLen, err = sConn.Write(buffer[2:payloadLen]); WLen <= 0 || err != nil {
		return
	}
	sConn.SetReadDeadline(time.Now().Add(config.Udp_timeout))
	if RLen, err = sConn.Read(buffer[2:]); RLen <= 0 || err != nil {
		return
	}
	//包长度转换
	buffer[0] = byte(RLen >> 8)
	buffer[1] = byte(RLen)
	//加密
	if len(CuteBi_XorCrypt_password) != 0 {
		CuteBi_XorCrypt(buffer[:2+RLen], 0)
	}
	cConn.Write(buffer[:2+RLen])
}

func Respond_HttpDNS(cConn net.Conn, header []byte) bool {
	var domainBegin string
	httpDNS_DomainSub := bytes.Index(header[:], []byte("dn="))
	if httpDNS_DomainSub < 0 {
		return false
	}
	if _, err := fmt.Sscanf(string(header[httpDNS_DomainSub+3:]), "%s", &domainBegin); err != nil {
		log.Println(err)
		return false
	}
	domain := strings.Split(domainBegin, "&")
	//log.Println("httpDNS domain: [" + domain[0] + "]")
	defer cConn.Close()
	ips, err := net.LookupHost(domain[0])
	if err != nil {
		cConn.Write([]byte("HTTP/1.0 404 Not Found\r\nConnection: Close\r\nServer: CuteBi Linux Network httpDNS, (%>w<%)\r\nContent-type: charset=utf-8\r\n\r\n<html><head><title>HTTP DNS Server</title></head><body>查询域名失败<br/><br/>By: 萌萌萌得不要不要哒<br/>E-mail: 915445800@qq.com</body></html>"))
		//log.Println("httpDNS domain: [" + domain[0] + "], Lookup failed")
	} else {
		isIpv6 := bytes.Contains(header[:], []byte("type=AAAA"))
		for i := 0; i < len(ips); i++ {
			if strings.Contains(ips[i], ":") /*只有ipv6才包含':'*/ == isIpv6 {
				if bytes.Contains(header[:], []byte("ttl=1")) {
					ips[i] += ",60"
				}
				fmt.Fprintf(cConn, "HTTP/1.0 200 OK\r\nConnection: Close\r\nServer: CuteBi Linux Network httpDNS, (%%>w<%%)\r\nContent-Length: %d\r\n\r\n%s", len(ips[i]), ips[i])
				break
			}
		}
		//log.Println("httpDNS domain: ["+domain[0]+"], IPS: ", ips)
	}
	return true
}
