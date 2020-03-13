// cns.go
package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

var (
	listener                                          *net.TCPListener
	udpFlag                                           string
	proxyKey                                          []byte
	tcp_timeout, udp_timeout, tcp_keepAlive           time.Duration
	enable_dns_tcpOverUdp, enable_httpDNS, enable_TFO bool
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

func handleConn(cConn *net.TCPConn, payload []byte) {
	cConn.SetKeepAlive(true)
	cConn.SetKeepAlivePeriod(tcp_keepAlive)
	cConn.SetReadDeadline(time.Now().Add(tcp_timeout))

	RLen, err := cConn.Read(payload)
	if err != nil || RLen <= 0 {
		cConn.Close()
		return
	}
	if isHttpHeader(payload[:RLen]) == false {
		handleUdpSession(cConn, payload[:RLen])
	} else {
		if enable_httpDNS == false || Respond_HttpDNS(cConn, payload[:RLen]) == false { /*优先处理httpDNS请求*/
			if WLen, err := cConn.Write(rspHeader(payload[:RLen])); err != nil || WLen <= 0 {
				cConn.Close()
				return
			}
			if bytes.Contains(payload[:RLen], []byte(udpFlag)) == true {
				handleConn(cConn, payload) //httpUDP需要读取到二进制数据才进行处理
			} else {
				handleTcpSession(cConn, payload[:RLen])
			}
		}
	}
}

func pidSaveToFile(pidPath string) {
	fp, err := os.Create(pidPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	fp.WriteString(fmt.Sprintf("%d", os.Getpid()))
	if err != nil {
		fmt.Println(err)
	}
	fp.Close()
}

func handleCmd() {
	var listenAddrString, proxyKeyString, CuteBi_XorCrypt_passwordStr, pidPath string
	var help, enable_daemon bool

	flag.StringVar(&proxyKeyString, "proxy-key", "Host", "tcp request proxy host key")
	flag.StringVar(&udpFlag, "udp-flag", "httpUDP", "udp request flag string")
	flag.StringVar(&listenAddrString, "listen-addr", ":8989", "listen aaddress")
	flag.StringVar(&CuteBi_XorCrypt_passwordStr, "encrypt-password", "", "encrypt password")
	flag.Int64Var((*int64)(&tcp_timeout), "tcp-timeout", 600, "tcp timeout second")
	flag.Int64Var((*int64)(&udp_timeout), "udp-timeout", 30, "udp timeout second")
	flag.Int64Var((*int64)(&tcp_keepAlive), "tcp-keepalive", 60, "tcp keepalive second")
	flag.StringVar(&pidPath, "pid-path", "", "pid file path")
	flag.BoolVar(&enable_dns_tcpOverUdp, "dns-tcpOverUdp", false, "tcpDNS Over udpDNS switch")
	flag.BoolVar(&enable_httpDNS, "enable-httpDNS", false, "httpDNS server switch")
	flag.BoolVar(&enable_TFO, "enable-TFO", false, "listener tcpFastOpen switch")
	flag.BoolVar(&enable_daemon, "daemon", false, "daemon mode switch")
	flag.BoolVar(&help, "h", false, "")
	flag.BoolVar(&help, "help", false, "display this message")

	flag.Parse()
	if help == true {
		fmt.Println("　/) /)\n" +
			"ฅ(՞•ﻌ•՞)ฅ\n" +
			"CuteBi Network Server 0.1\nAuthor: CuteBi(Mmmdbybyd)\nE-mail: 915445800@qq.com\n")
		flag.Usage()
		os.Exit(0)
	}
	if enable_daemon == true {
		exec.Command(os.Args[0], []string(append(os.Args[1:], "-daemon=false"))...).Start()
		os.Exit(0)
	}
	listenAddr, err := net.ResolveTCPAddr("tcp", listenAddrString)
	listener, err = net.ListenTCP("tcp", listenAddr)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	if enable_TFO == true {
		enableTcpFastopen(listener)
	}
	if pidPath != "" {
		pidSaveToFile(pidPath)
	}
	proxyKey = []byte("\n" + proxyKeyString + ": ")
	CuteBi_XorCrypt_password = []byte(CuteBi_XorCrypt_passwordStr)
	tcp_timeout *= time.Second
	udp_timeout *= time.Second
	tcp_keepAlive *= time.Second
}

func initProcess() {
	handleCmd()
	setsid()
	setMaxNofile()
	signal.Ignore(syscall.SIGPIPE)
}

func main() {
	initProcess()
	runtime.GOMAXPROCS(runtime.NumCPU())
	defer listener.Close()

	var (
		conn *net.TCPConn
		err  error
	)
	for {
		conn, err = listener.AcceptTCP()
		if err != nil {
			log.Println(err)
			time.Sleep(3 * time.Second)
			continue
		}
		go handleConn(conn, make([]byte, 8192))
	}
}
