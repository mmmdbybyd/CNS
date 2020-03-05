// udp.go
package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

type UdpSession struct {
	cConn                                      *net.TCPConn
	udpSConn                                   *net.UDPConn
	decrypt_dynamic_code, encrypt_dynamic_code byte
}

func (udpSess *UdpSession) udpServerToClient() {
	defer udpSess.cConn.Close()
	defer udpSess.udpSConn.Close()

	/* 不要在for里用:=申请变量, 否则每次循环都会重新申请内存 */
	var (
		RAddr                              *net.UDPAddr
		payload_len, ignore_head_len, WLen int
		err                                error
	)
	payload := make([]byte, 65535)
	for {
		udpSess.udpSConn.SetReadDeadline(time.Now().Add(udp_timeout))
		payload_len, RAddr, err = udpSess.udpSConn.ReadFromUDP(payload[24:] /*24为httpUDP协议头保留使用*/)
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println("readUdpServerLen: ", payload_len, "RAddr: ", RAddr.String())
		if bytes.HasPrefix(RAddr.IP, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xff, 0xff}) == true {
			/* ipv4 */
			ignore_head_len = 12                 //数组前面的12字节不需要
			payload[12] = byte(payload_len + 10) //从第13个字节开始设置协议头
			payload[13] = byte((payload_len + 10) >> 8)
			copy(payload[14:18], []byte{0, 0, 0, 1})
			copy(payload[18:22], []byte(RAddr.IP)[12:16])
		} else {
			/* ipv6 */
			ignore_head_len = 0
			payload[0] = byte(payload_len + 22)
			payload[1] = byte((payload_len + 22) >> 8)
			copy(payload[2:6], []byte{0, 0, 0, 3})
			copy(payload[6:22], []byte(RAddr.IP))
		}
		payload[22] = byte(RAddr.Port >> 8)
		payload[23] = byte(RAddr.Port)
		if passLen_code != 0 {
			udpSess.encrypt_dynamic_code = CuteBi_encrypt(payload[ignore_head_len:24+payload_len], udpSess.encrypt_dynamic_code)
		}
		udpSess.cConn.SetReadDeadline(time.Now().Add(udp_timeout))
		if WLen, _ = udpSess.cConn.Write(payload[ignore_head_len : 24+payload_len]); WLen <= 0 {
			return
		}
	}
}

func (udpSess *UdpSession) writeToServer(httpUDP_data []byte) int {
	var (
		udpAddr                                   net.UDPAddr
		WLen                                      int
		pkgSub, pkgLen, httpUDP_protocol_head_len uint16
	)
	for pkgSub = 0; (pkgSub + 2) < uint16(len(httpUDP_data)); pkgSub += 2 + pkgLen {
		pkgLen = uint16(httpUDP_data[pkgSub]) | (uint16(httpUDP_data[pkgSub+1]) << 8) //2字节储存包的长度，包括socks5头
		log.Println("pkgLen: ", pkgLen, "  ", uint16(len(httpUDP_data)))
		if pkgSub+pkgLen > uint16(len(httpUDP_data)) || pkgLen <= 12 {
			return 0
		}
		if httpUDP_data[5] == 1 {
			/* ipv4 */
			udpAddr.IP = net.IPv4(httpUDP_data[pkgSub+6], httpUDP_data[pkgSub+7], httpUDP_data[pkgSub+8], httpUDP_data[pkgSub+9])
			udpAddr.Port = int((uint16(httpUDP_data[pkgSub+10]) << 8) | uint16(httpUDP_data[pkgSub+11]))
			httpUDP_protocol_head_len = 12
		} else {
			if pkgLen <= 24 {
				return 0
			}
			/* ipv6 */
			udpAddr.IP = net.IP(httpUDP_data[pkgSub+6 : pkgSub+22])
			udpAddr.Port = int((uint16(httpUDP_data[pkgSub+22]) << 8) | uint16(httpUDP_data[pkgSub+23]))
			httpUDP_protocol_head_len = 24
		}
		log.Println("WriteToUdpAddr: ", udpAddr.String())
		WLen, _ = udpSess.udpSConn.WriteToUDP(httpUDP_data[pkgSub+httpUDP_protocol_head_len:pkgSub+2+pkgLen], &udpAddr)
		if WLen <= 0 {
			return -1
		}
	}

	return int(pkgSub)
}

func (udpSess *UdpSession) udpClientToServer(httpUDP_data []byte) {
	defer udpSess.cConn.Close()
	defer udpSess.udpSConn.Close()

	var payload_len, RLen, WLen int
	WLen = udpSess.writeToServer(httpUDP_data)
	if WLen == -1 {
		return
	}
	payload := make([]byte, 65535)
	if WLen < len(httpUDP_data) {
		payload_len = copy(payload, httpUDP_data[WLen:])
	}
	for {
		udpSess.cConn.SetReadDeadline(time.Now().Add(udp_timeout))
		RLen, _ = udpSess.cConn.Read(payload[payload_len:])
		if RLen <= 0 {
			return
		}
		if passLen_code != 0 {
			udpSess.decrypt_dynamic_code = CuteBi_decrypt(payload[payload_len:payload_len+RLen], udpSess.decrypt_dynamic_code)
		}
		payload_len += RLen
		WLen = udpSess.writeToServer(payload[:payload_len])
		if WLen == -1 {
			return
		} else if WLen < len(payload[:payload_len]) {
			payload_len = copy(payload, payload[WLen:payload_len])
		} else {
			payload_len = 0
		}
	}
}

func (udpSess *UdpSession) initUdp(httpUDP_data []byte) bool {
	if passLen_code != 0 {
		var decryptSuccess bool
		if udpSess.decrypt_dynamic_code, decryptSuccess = verify_CuteBiEncrypt_dynamic_code(httpUDP_data[:3]); decryptSuccess == false {
			return false
		}
		udpSess.encrypt_dynamic_code = udpSess.decrypt_dynamic_code
		udpSess.decrypt_dynamic_code = CuteBi_decrypt(httpUDP_data, udpSess.decrypt_dynamic_code)
	}
	var err error
	udpSess.udpSConn, err = net.ListenUDP("udp", nil)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func handleUdpSession(cConn *net.TCPConn, httpUDP_data []byte) {
	udpSess := new(UdpSession)

	defer func() {
		log.Println("A udp client close")
		cConn.Close()
	}()

	udpSess.cConn = cConn
	if udpSess.initUdp(httpUDP_data) == false {
		log.Println("Is not httpUDP protocol")
		return
	}
	log.Println("Start udpForward")
	go udpSess.udpClientToServer(httpUDP_data)
	udpSess.udpServerToClient()
}

func Respond_HttpDNS(cConn *net.TCPConn, header []byte) bool {
	var domain string
	httpDNS_DomainSub := bytes.Index(header[:], []byte("?dn="))
	if httpDNS_DomainSub < 0 {
		return false
	}
	if _, err := fmt.Sscanf(string(header[httpDNS_DomainSub+4:]), "%s", &domain); err != nil {
		log.Println(err)
		return false
	}
	log.Println("httpDNS domain: [" + domain + "]")
	defer cConn.Close()
	ips, err := net.LookupHost(domain)
	if err != nil {
		cConn.Write([]byte("HTTP/1.0 404 Not Found\r\nConnection: Close\r\nServer: CuteBi Linux Network httpDNS, (%>w<%)\r\nContent-type: charset=utf-8\r\n\r\n<html><head><title>HTTP DNS Server</title></head><body>查询域名失败<br/><br/>By: 萌萌萌得不要不要哒<br/>E-mail: 915445800@qq.com</body></html>"))
		log.Println("httpDNS domain: [" + domain + "], Lookup failed")
	} else {
		for i := 0; i < len(ips); i++ {
			if strings.Contains(ips[i], ":") == false { //跳过ipv6
				fmt.Fprintf(cConn, "HTTP/1.0 200 OK\r\nConnection: Close\r\nServer: CuteBi Linux Network httpDNS, (%%>w<%%)\r\nContent-Length: %d\r\n\r\n%s", len(string(ips[i])), string(ips[i]))
				break
			}
		}
		log.Println("httpDNS domain: ["+domain+"], IPS: ", ips)
	}
	return true
}
