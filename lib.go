package well_of_file

import (
	"fmt"
	"github.com/Gaojianli/well_of_file/config"
	"github.com/Gaojianli/well_of_file/stage"
	"net"
	"os"
)

func Send(fileInput string, port int) {
	address := fmt.Sprintf("%s:%d", config.BIND_ADDRESS, port)
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()
	fileInfo, err := os.Stat(fileInput)
	if os.IsNotExist(err) {
		fmt.Println("file not exist!")
		os.Exit(-1)
	} else if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	fmt.Printf("Server is listening %s\n", address)
	remote, err := stage.HandShakeServer(conn, fileInfo)
	if err != nil {
		println("Handshake failed!")
		println(err.Error())
		os.Exit(-1)
	}
	// start send file
	err = stage.Send(conn, remote, fileInput)
	if err != nil {
		println("SendFile failed!")
		println(err.Error())
		os.Exit(-1)
	}
}

func Receive(serverAddress string, port int, saveTo string) {
	address := fmt.Sprintf("%s:%d", serverAddress, port)
	conn, err := net.Dial("udp", address)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()
	meta, err := stage.HandShakeClient(conn)
	if err != nil {
		println("Handshake failed!")
		println(err.Error())
		os.Exit(-1)
	}
	fmt.Printf("Succeed handshake, file name: %s\n", meta.FileName)
	err = stage.Receive(conn, meta, saveTo)
	if err != nil {
		println("Receive failed!")
		println(err.Error())
		os.Exit(-1)
	}
}
