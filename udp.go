package main

import (
	"bytes"
	"log"
	"net"
	"time"
)

type UdpSession struct {
	cConn                                                            net.Conn
	udpSConn                                                         *net.UDPConn
	c2s_CuteBi_XorCrypt_passwordSub, s2c_CuteBi_XorCrypt_passwordSub int
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
	payload := make([]byte, 65536)
	for {
		udpSess.cConn.SetReadDeadline(time.Now().Add(config.Udp_timeout))
		udpSess.udpSConn.SetReadDeadline(time.Now().Add(config.Udp_timeout))
		payload_len, RAddr, err = udpSess.udpSConn.ReadFromUDP(payload[24:] /*24为httpUDP协议头保留使用*/)
		if err != nil || payload_len <= 0 {
			return
		}
		//fmt.Println("readUdpServerLen: ", payload_len, "RAddr: ", RAddr.String())
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
		if len(CuteBi_XorCrypt_password) != 0 {
			udpSess.s2c_CuteBi_XorCrypt_passwordSub = CuteBi_XorCrypt(payload[ignore_head_len:24+payload_len], udpSess.s2c_CuteBi_XorCrypt_passwordSub)
		}
		udpSess.cConn.SetWriteDeadline(time.Now().Add(config.Udp_timeout))
		if WLen, err = udpSess.cConn.Write(payload[ignore_head_len : 24+payload_len]); err != nil || WLen <= 0 {
			return
		}
	}
}

func (udpSess *UdpSession) writeToServer(httpUDP_data []byte) int {
	var (
		udpAddr                           net.UDPAddr
		err                               error
		WLen                              int
		pkgSub, httpUDP_protocol_head_len int
		pkgLen                            uint16
	)
	for pkgSub = 0; pkgSub+2 < len(httpUDP_data); pkgSub += 2 + int(pkgLen) {
		pkgLen = uint16(httpUDP_data[pkgSub]) | (uint16(httpUDP_data[pkgSub+1]) << 8) //2字节储存包的长度，包括socks5头
		//log.Println("pkgSub: ", pkgSub, ", pkgLen: ", pkgLen, "  ", uint16(len(httpUDP_data)))
		if pkgSub+2+int(pkgLen) > len(httpUDP_data) || pkgLen <= 10 {
			return 0
		}
		if bytes.HasPrefix(httpUDP_data[pkgSub+3:pkgSub+5], []byte{0, 0}) == false {
			return 1
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
		//log.Println("WriteToUdpAddr: ", udpAddr.String())
		if WLen, err = udpSess.udpSConn.WriteToUDP(httpUDP_data[pkgSub+httpUDP_protocol_head_len:pkgSub+2+int(pkgLen)], &udpAddr); err != nil || WLen <= 0 {
			return -1
		}
	}

	return int(pkgSub)
}

func (udpSess *UdpSession) udpClientToServer(httpUDP_data []byte) {
	defer udpSess.cConn.Close()
	defer udpSess.udpSConn.Close()

	var payload_len, RLen, WLen int
	var err error
	WLen = udpSess.writeToServer(httpUDP_data)
	if WLen == -1 {
		return
	}
	payload := make([]byte, 65536)
	if WLen < len(httpUDP_data) {
		payload_len = copy(payload, httpUDP_data[WLen:])
	}
	for {
		udpSess.cConn.SetReadDeadline(time.Now().Add(config.Udp_timeout))
		udpSess.udpSConn.SetReadDeadline(time.Now().Add(config.Udp_timeout))
		RLen, err = udpSess.cConn.Read(payload[payload_len:])
		if err != nil || RLen <= 0 {
			return
		}
		if len(CuteBi_XorCrypt_password) != 0 {
			udpSess.c2s_CuteBi_XorCrypt_passwordSub = CuteBi_XorCrypt(payload[payload_len:payload_len+RLen], udpSess.c2s_CuteBi_XorCrypt_passwordSub)
		}
		payload_len += RLen
		//log.Println("Read Client: ", payload_len)
		WLen = udpSess.writeToServer(payload[:payload_len])
		if WLen == -1 {
			return
		} else if WLen < payload_len {
			payload_len = copy(payload, payload[WLen:payload_len])
		} else {
			payload_len = 0
		}
	}
}

func (udpSess *UdpSession) initUdp(httpUDP_data []byte) bool {
	if len(CuteBi_XorCrypt_password) != 0 {
		de := make([]byte, 5)
		copy(de, httpUDP_data[0:5])
		CuteBi_XorCrypt(de, 0)
		if de[2] != 0 || de[3] != 0 || de[4] != 0 {
			return false
		}
		udpSess.c2s_CuteBi_XorCrypt_passwordSub = CuteBi_XorCrypt(httpUDP_data, 0)
	}
	var err error
	udpSess.udpSConn, err = net.ListenUDP("udp", nil)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

func handleUdpSession(cConn net.Conn, httpUDP_data []byte) {
	//defer log.Println("A udp client close")

	udpSess := new(UdpSession)
	udpSess.cConn = cConn
	if udpSess.initUdp(httpUDP_data) == false {
		cConn.Close()
		log.Println("Is not httpUDP protocol or Decrypt failed")
		return
	}
	//log.Println("Start udpForward")
	go udpSess.udpClientToServer(httpUDP_data)
	udpSess.udpServerToClient()
}
