// tlsHandle
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"time"
)

type TlsServer struct {
	Listen_addr, AutoCertHosts []string
	CertFile, KeyFile          string
}

func createSSLcertificate(hosts string) ([]byte, []byte) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("Failed to generate serial number: %v", err)
	}
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, h := range hosts {
		/*if ip := net.ParseIP(string(h)); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {*/
		template.DNSNames = append(template.DNSNames, string(h))
		//}
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		log.Fatalf("Failed to create certificate: %v", err)
	}
	keyBytes, _ := x509.MarshalPKCS8PrivateKey(priv)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	return certPEM, keyPEM
}

func (tlsConfig *TlsServer) startTls() {
	certs := make([]tls.Certificate, 0)
	if tlsConfig.CertFile != "" && tlsConfig.KeyFile != "" {
		cer, err := tls.LoadX509KeyPair(tlsConfig.CertFile, tlsConfig.KeyFile)
		if err != nil {
			log.Println(err)
			return
		}
		certs = append(certs, cer)
	}
	if tlsConfig.AutoCertHosts != nil {
		for _, h := range tlsConfig.AutoCertHosts {
			cer, err := tls.X509KeyPair(createSSLcertificate(h))
			if err != nil {
				log.Println(err)
				return
			}
			certs = append(certs, cer)
		}
	}

	handleFun := func(listenAddrString string) {
		listener, err := tls.Listen("tcp", listenAddrString, &tls.Config{Certificates: certs})
		if err != nil {
			log.Println(err)
			return
		}
		defer listener.Close()

		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println(err)
				time.Sleep(3 * time.Second)
				continue
			}
			go handleConn(conn, make([]byte, 8192))
		}
	}
	for i := len(tlsConfig.Listen_addr) - 1; i > 0; i-- {
		go handleFun(tlsConfig.Listen_addr[i])
	}
	handleFun(tlsConfig.Listen_addr[0])
}
