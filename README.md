# CuteBi Network Server (CNS)

基于 Go 实现的网络代理服务端，支持 IPv6、TCP FastOpen（Windows 暂不支持）、UDP over HTTP 隧道；需配合专用客户端（如 [CLNC](https://github.com/mmmdbybyd/CLNC)）可实现 TCP/UDP 全局代理。

---

## 实际功能（按代码实现）

### 1. HTTP CONNECT / 隧道代理

- 监听 `listen_addr` 上的 TCP 连接，识别 HTTP 请求头（CONNECT、GET、POST 等）。
- 从请求头中按 `proxy_key`（默认 `Host`）取目标 host，建立到目标地址的 TCP 连接，双向转发数据。
- 支持可选 XOR 流式加密（`encrypt_password`）；若开启，Host 为 Base64 + XOR 加密后放在请求头中。

### 2. HTTP DNS 服务端

- 请求uri参数中带 `dn=域名` 时，走 HTTP DNS 逻辑：对本机做 `net.LookupHost(domain)`，返回 IP。
- 参数支持 `type=AAAA` 返回 IPv6，`ttl=1` 在响应中带 TTL。
- 与 114DNS、腾讯 DNSPod 等 HTTP DNS 用法类似，本机解析、HTTP 响应。

### 3. TCP DNS 转 UDP DNS

- 当目标 host 为 `*:53` 时，将客户端发来的 TCP DNS（2 字节长度 + 载荷）转为对同一 host 的 UDP DNS 请求，收到 UDP 应答后再按 TCP 长度前缀写回客户端。
- 由配置项 `Enable_dns_tcpOverUdp` 控制，可解决部分DNS服务器不支持TCPDNS的问题。

### 4. UDP over HTTP 隧道（httpUDP）

- 请求头中包含 `Udp_flag`（默认 `httpUDP`）时，该连接被当作 UDP 会话处理。
- 协议格式：2 字节包长度 + 协议头（IPv4 约 12 字节 / IPv6 约 24 字节）+ UDP 载荷；服务端解析后对目标 IP:Port 做 UDP 收发，再按同格式回写客户端。
- 支持 IPv4 / IPv6；可选 XOR 加密（与 TCP 隧道共用 `encrypt_password`）。

### 5. TLS 与 TCP FastOpen

- **TLS**：配置 `Tls.listen_addr`，在这些地址上先做 TLS 服务端握手，再按上述同一套隧道逻辑处理（HTTP DNS / CONNECT / httpUDP）。
- **证书**：支持 `CertFile`/`KeyFile` 或 `AutoCertHosts`（按 host 自动生成 ECDSA P256 证书）。
- **TCP FastOpen**：由 `Enable_TFO` 控制，仅在非 Windows 且进程有效 UID 为 0 时生效；Windows 下不启用。

---

## 命令行参数说明

| 参数 | 说明 |
|--------|------|
| `h` / `help` | 显示帮助 |
| `json` | json配置文件路径 |
| `daemon` | 值为true时开启守护进程 |

---

## 配置文件说明

| 配置项 | 说明 |
|--------|------|
| `Listen_addr` | 字符串数组，监听地址 |
| `Tcp_timeout` / `Udp_timeout` | 数值，TCP、UDP 超时（单位：秒） |
| `Proxy_key` | 字符串，从请求头取目标地址的 key，默认 `Host` |
| `Udp_flag` | 字符串，HTTP请求头中标识 httpUDP 的字符串，默认 `httpUDP` |
| `Encrypt_password` | 字符串， XOR 加密密码，空则不加密 |
| `Enable_dns_tcpOverUdp` | 布尔，是否开启 TCP DNS 转 UDP DNS |
| `Enable_httpDNS` | 布尔，是否开启 HTTP DNS（默认 `true`） |
| `Enable_TFO` | 布尔，是否开启 TCP FastOpen（仅 Linux + root 生效） |
| `Pid_path` | 字符串，可选，写入 PID 的文件路径 |
| `Tls.Listen_addr` | 字符串数组，TLS监听地址 |
| `Tls.AutoCertHosts` | 字符串数组，按 host 数组生成自签名的TLS证书 |
| `Tls.CertFile` | 字符串，本地TLS证书文件路径 |
| `Tls.KeyFile` | 字符串，本地TLS密钥文件路径 |

示例配置见 [config/cns.json](config/cns.json)。

---

## 编译与运行

```bash
go build -ldflags="-w -s" -trimpath -o cns
./cns -daemon=true -json=cns.json
```

---

## Linux一键:  
~~~~~
安装: `type curl &>/dev/null && echo 'curl -O' || echo 'wget -O cns.sh'` http://binary.parso.org/cns/cns.sh && sh cns.sh  
卸载: `type curl &>/dev/null && echo 'curl -O' || echo 'wget -O cns.sh'` http://binary.parso.org/cns/cns.sh && sh cns.sh uninstall  
~~~~~

---

## 项目结构（与源码对应）

| 文件 | 职责 |
|------|------|
| `cns.go` | 入口、命令行解析、JSON 配置加载、守护进程与 TLS/HTTP 隧道启动 |
| `http_tunnel.go` | TCP 监听、首包判断 HTTP/非 HTTP、回复隧道头、分流到 HTTP DNS / TCP 会话 / UDP 会话，可选 TLS 包装 |
| `tcp.go` | 从请求头取 host、TCP 转发、`:53` 时转 dns_tcpOverUdp |
| `udp.go` | httpUDP 协议解析、UDP 收发、IPv4/IPv6 头与 XOR 加解密 |
| `dns.go` | HTTP DNS 响应（`dn=`）、TCP DNS 转 UDP DNS（`dns_tcpOverUdp`） |
| `tlsSide.go` | TLS 监听、证书加载/自动生成、TLS 包装后交给 handleTunnel |
| `CuteBi_XorCrypt.go` | XOR 流式加解密、Host 的 Base64+XOR 解密 |
| `sys_isWin.go` / `sys_isNotWin.go` | Windows 与 非 Windows 的 TFO、setMaxNofile、setsid |

---
