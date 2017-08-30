package proxy
import(
	"net"
	"github.com/crabkun/DazeClient/common"
	"log"
)
func StartProxy() (error){
	listener, err := net.Listen("tcp", "127.0.0.1:"+common.SrvConf.LocalPort)
	if err != nil {
		log.Println("本地HTTP/SOCKS5代理监听失败！原因： ", err)
		return err
	}
	log.Println("本地HTTP/SOCKS5代理成功监听于",listener.Addr())
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn)
	}
	return nil
}
func handleConnection(conn net.Conn){
	testchar:=make([]byte,1)
	_,err:=conn.Read(testchar)
	if err!=nil{
		return
	}
	if testchar[0]==5{
		Socks5handleConnection(&SwitchConn{conn,testchar})
	}else{
		HTTPProxyHandle(&SwitchConn{conn,testchar})
	}
}

type SwitchConn struct {
	net.Conn
	Testchar []byte
}
func (this *SwitchConn) Read(b []byte) (n int, err error){
	if this.Testchar!=nil{
		b[0]=this.Testchar[0]
		this.Testchar=nil
		b=b[1:]
		n,err:=this.Conn.Read(b)
		return n+1,err
	}
	 return this.Conn.Read(b)
}