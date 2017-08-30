package proxy

import (
	"net/http"
	"net"
	"bufio"
	"bytes"
	"github.com/crabkun/DazeClient/common"
	"github.com/crabkun/DazeClient/server"
	"regexp"
)
//func StartHttpProxy()(error){
//	l,err:=net.Listen("tcp","127.0.0.1:"+common.SrvConf.HTTPProxyPort)
//	if err!=nil{
//		log.Println("本地HTTP(S)代理监听失败！原因：",err.Error())
//		return err
//	}
//	log.Println("本地HTTP(S)代理成功监听在",l.Addr())
//	for {
//		conn,err:=l.Accept()
//		if err!=nil{
//			continue
//		}
//		go ProxyHandle(conn)
//	}
//	return nil
//}
func HTTPProxyHandle(c net.Conn){
	proto:="http"
	defer func(){
		c.Close()
	}()
	var err error
	//解析http代理数据包
	r:=bufio.NewReader(c)
	rq,err:=http.ReadRequest(r)
	if err!=nil{
		return
	}
reconnect:
	if rq.Method=="CONNECT"{
		proto="https"
	}
	host:=rq.Host
	if b,_:=regexp.MatchString("^.+:[0-9]+$",rq.Host);!b{
		if proto=="http"{
			host=rq.Host+":80"
		}else{
			host=rq.Host+":443"
		}
	}
	//呼叫代理服务器
	client:=server.CallProxyServer(c,common.SrvConf,host,"tcp")
	if client==nil{
		return
	}
	defer func(){
		client.RemoteServerConn.Close()
	}()

	if proto=="https"{
		//特殊处理https代理客户端
		c.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	}else{
		rq.Write(client)
	}

	//IO桥：代理服务器到客户端
	go func(client *server.S_Client){
		defer func(){
			recover()
			client.ProxyUserConn.Close()
		}()
		for{
			b:=client.Read()
			client.SafeSend(b,client.ProxyUserConn)
		}
	}(client)

	//IO桥：客户端到代理服务器
	if rq=func(client *server.S_Client) *http.Request{
		buf:=make([]byte,65500)
		for{
			n,err:=client.ProxyUserConn.Read(buf)
			if err!=nil{
				return nil
			}
			if proto=="http"{
				if nrq:=IsHTTPpacket(buf[:n]);nrq!=nil{
					return nrq
				}
			}
			client.Write(buf[:n])
		}
	}(client);rq!=nil{
		goto reconnect
	}

}
func IsHTTPpacket(buf []byte) *http.Request{
	r:=bufio.NewReader(bytes.NewReader(buf))
	req,ReadRequestErr:=http.ReadRequest(r)
	if ReadRequestErr!=nil{
		return nil
	}
	return req
}