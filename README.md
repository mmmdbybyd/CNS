# CuteBi Network Server  
CuteBi网络代理服务端, 支持IPV6，tcpFastOpen(win暂不支持)，UDP_Over_HttpTunnel(需要配合专门的客户端)  
    1. 普通的CONNECT代理服务器(暂时不考虑添加普通http支持)  
    2. 实现与114DNS以及腾讯的dnsPod一样的httpDNS服务端  
    3. 配合专门的客户端可以实现TCP/UDP全局代理, 目前有: [CLNC](https://github.com/mmmdbybyd/CLNC)
  
单独服务端:  
--------
    1. 普通的CONNECT代理服务器(暂时不考虑添加普通http支持)  
    2. 实现与114DNS以及腾讯的dnsPod一样的httpDNS服务端  
  
服务端+客户端:
--------
    1. 可伪装为各种HTTP/HTTPS数据, 并加密传输流量(可选)  
    2. 支持UDP_Over_HttpTunnel  
    3. 支持tcpDNS转udpDNS解析dns  
  
##### BUG:  
&nbsp;&nbsp;&nbsp;&nbsp;/) /)  
ฅ(• ﻌ •)ฅ  
暂无发现bug  
  
##### 编译:  
~~~~~
go build -o cns  
~~~~~
  
##### 启动命令:  
[配置文件格式](config/cns.json)
~~~~~
./cns -daemon=true -json=cns.json
~~~~~
  
##### Linux一键:  
~~~~~
安装: `type curl &>/dev/null && echo 'curl -O' || echo 'wget -O cns.sh'` http://pros.cutebi.taobao69.cn:666/cns/cns.sh && sh cns.sh  
卸载: `type curl &>/dev/null && echo 'curl -O' || echo 'wget -O cns.sh'` http://pros.cutebi.taobao69.cn:666/cns/cns.sh && sh cns.sh uninstall  
~~~~~
