package stage

import (
	"github.com/tmthrgd/go-memset"
	"github.com/Gaojianli/well_of_file/config"
	protocol "github.com/Gaojianli/well_of_file/protobuf/idl"
	"net"
	"os"
)

func HandShakeServer(conn *net.UDPConn, fileinfo os.FileInfo) (*net.UDPAddr, error) {
	buff := make([]byte, 500)
handshake:
	count, clientSock, err := conn.ReadFromUDP(buff)
	if err != nil {
		return nil, err
	}
	helloMsg := protocol.Hello{}
	err = helloMsg.Unmarshal(buff[:count])
	if err != nil {
		println("Failed to handshake, try again")
		goto handshake
	}
	metaMsg := protocol.Meta{
		FileName:    fileinfo.Name(),
		Length:      fileinfo.Size(),
		ChunkSize:   config.CHUNK_SZIE,
		PackageSize: config.PACKAGE_SIZE,
	}
	metaStr, _ := metaMsg.Marshal()
	for {
		_, _ = conn.WriteToUDP(metaStr, clientSock)
		memset.Memset(buff, 0)
		count, clientSock, err = conn.ReadFromUDP(buff)
		recvMeta := protocol.Meta{}
		err = recvMeta.Unmarshal(buff[:count])
		if err == nil {
			if recvMeta.FileName == metaMsg.FileName {
				break
			}
		}
	}
	return clientSock, nil
}

func HandShakeClient(conn  net.Conn) (protocol.Meta,error){
	meta := protocol.Meta{}
	hostname,err :=os.Hostname()
	if err!=nil{
		return meta,err
	}
	helloMsg := protocol.Hello{
		Hostname: hostname,
	}
handshake:
	helloStr,_ :=helloMsg.Marshal()
	buffer := make([]byte,500)
	_,_ = conn.Write(helloStr)
	count,err:=conn.Read(buffer)
	if err!=nil{
		return meta,err
	}
	err = meta.Unmarshal(buffer[:count])
	if err!=nil{
		println("Failed to handshake, try again")
		goto handshake
	}
	metaStr,_:=meta.Marshal()
	_, _ = conn.Write(metaStr)
	return meta,nil
}