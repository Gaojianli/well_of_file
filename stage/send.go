package stage

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"github.com/Gaojianli/ltcode/codec"
	"github.com/Gaojianli/ltcode/utils"
	"github.com/Gaojianli/well_of_file/config"
	protocol "github.com/Gaojianli/well_of_file/protobuf/idl"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
)

func Send(conn *net.UDPConn, remote *net.UDPAddr, filepath string) error {
	fs, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer fs.Close()
	r := bufio.NewReader(fs)
	fileBuffer := make([]byte, config.CHUNK_SZIE)
	eof := false
	encoder := codec.Codec{}
	encoder.Init(config.CHUNK_SZIE/config.PACKAGE_SIZE+1, utils.SolitonDistribution(config.CHUNK_SZIE/config.PACKAGE_SIZE+1))
	chunkId := 0
	for {
		if eof {
			println("All chunks sended..")
			break
		}
		count, err := r.Read(fileBuffer)
		isFirst := false
		if err == io.EOF {
			eof = true
		} else if err != nil {
			log.Fatal(err)
		}
		isFin := false
		for !isFin {
			encodeRes := encoder.EncodeBlocks(fileBuffer, []int64{rand.Int63()})
			pack := protocol.Package{
				ChunkId: int64(chunkId),
				BlockId: encodeRes[0].BlockCode,
				Length:  int64(count),
				Data:    []byte(base64.StdEncoding.EncodeToString(encodeRes[0].Data)),
			}
			// 监听停止
			if !isFirst {
				go func() {
					finBuf := make([]byte, 50)
					for {
						finCount, _, err := conn.ReadFromUDP(finBuf)
						if err != nil {
							log.Fatal(err)
						}
						fin := protocol.Fin{}
						if fin.Unmarshal(finBuf[:finCount]) == nil {
							fmt.Printf("[Chunk %d]: Chunk %d sended.\n", fin.ChunkId, fin.ChunkId)
							if fin.ChunkId == int64(chunkId) {
								isFin = true
								break
							}
						}
					}
				}()
				isFirst = true
			}
			packStr, _ := pack.Marshal()
			_, _ = conn.WriteToUDP(packStr, remote)
			fmt.Printf("[Chunk %d]: Package %d sended...\n", chunkId, encodeRes[0].BlockCode)
			//time.Sleep(time.Second)
		}
		chunkId++
	}
	return nil
}
