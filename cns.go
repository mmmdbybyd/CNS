package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

type JsonConfig struct {
	Tls                                               TlsServer
	Listen_addr                                       []string
	Proxy_key, Udp_flag, Encrypt_password, Pid_path   string
	Tcp_timeout, Udp_timeout                          time.Duration
	Enable_dns_tcpOverUdp, Enable_httpDNS, Enable_TFO bool
}

var config = JsonConfig{
	Proxy_key:   "Host",
	Udp_flag:    "httpUDP",
	Tcp_timeout: 600,
	Udp_timeout: 30,
}

func jsonLoad(filename string, v *JsonConfig) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = json.Unmarshal(data, v)
	if err != nil {
		log.Fatal(err)
		return
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
	var (
		err                 error
		jsonConfigPath      string
		help, enable_daemon bool
	)

	flag.StringVar(&jsonConfigPath, "json", "", "json config path")
	flag.BoolVar(&enable_daemon, "daemon", false, "daemon mode switch")
	flag.BoolVar(&help, "h", false, "")
	flag.BoolVar(&help, "help", false, "display this message")

	flag.Parse()
	if help == true {
		fmt.Println("　/) /)\n" +
			"ฅ(՞•ﻌ•՞)ฅ\n" +
			"CuteBi Network Server 0.3.2\nAuthor: CuteBi(Mmmdbybyd)\nE-mail: 915445800@qq.com\n")
		flag.Usage()
		os.Exit(0)
	}
	if jsonConfigPath == "" {
		flag.Usage()
		fmt.Println("\n\nFind't json config file")
		os.Exit(1)
	}
	if enable_daemon == true {
		exec.Command(os.Args[0], []string(append(os.Args[1:], "-daemon=false"))...).Start()
		os.Exit(0)
	}
	jsonLoad(jsonConfigPath, &config)

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	//有效uid不为0(root)的关闭tfo
	if config.Enable_TFO == true && os.Geteuid() != 0 {
		config.Enable_TFO = false
		fmt.Println("TFO cannot be opened: CNS effective UID isn't 0(root).")
	}
	if config.Pid_path != "" {
		pidSaveToFile(config.Pid_path)
	}
	config.Enable_httpDNS = true
	config.Proxy_key = "\n" + config.Proxy_key + ": "
	CuteBi_XorCrypt_password = []byte(config.Encrypt_password)
	config.Tcp_timeout *= time.Second
	config.Udp_timeout *= time.Second
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
	if config.Tls.AutoCertHosts != nil || (config.Tls.CertFile != "" && config.Tls.KeyFile != "") {
		config.Tls.makeCertificateConfig()
		for i := len(config.Tls.Listen_addr) - 1; i >= 0; i-- {
			go config.Tls.startTls(config.Tls.Listen_addr[i])
		}
	}
	for i := len(config.Listen_addr) - 1; i >= 0; i-- {
		go startHttpTunnel(config.Listen_addr[i])
	}
	select {}
}
