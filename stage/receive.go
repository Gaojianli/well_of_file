package stage

import (
	"github.com/Gaojianli/ltcode/codec"
	ltutils "github.com/Gaojianli/ltcode/utils"
	"encoding/base64"
	"fmt"
	"github.com/Gaojianli/well_of_file/config"
	protocol "github.com/Gaojianli/well_of_file/protobuf/idl"
	"github.com/Gaojianli/well_of_file/utils"
	"net"
	"sync"
)

type chunkStatus struct {
	decoder *codec.Decoder
	done    bool
}

func Receive(conn net.Conn, meta protocol.Meta, saveTo string) error {
	codecMap := make(map[int64]chunkStatus)
	buffer := make([]byte, config.PACKAGE_SIZE*1.4+50) // base64体积最多膨胀这么多
	writeFileWg := sync.WaitGroup{}
	isFin := false
	for !isFin{
		count, _ := conn.Read(buffer)
		pack := protocol.Package{}
		if pack.Unmarshal(buffer[:count]) == nil {
			rawData, err := base64.StdEncoding.DecodeString(string(pack.Data))
			if err != nil {
				println("Failed to parse package")
				println(err.Error())
				continue
			}
			if _, ok := codecMap[pack.ChunkId]; !ok {
				c := codec.Codec{}
				c.Init(
					config.CHUNK_SZIE/config.PACKAGE_SIZE + 1,
					ltutils.SolitonDistribution(config.CHUNK_SZIE/config.PACKAGE_SIZE + 1),
				)
				codecMap[pack.ChunkId] = chunkStatus{
					decoder: c.GetDecoder(config.CHUNK_SZIE),
					done:    false,
				}
			}
			decoder, _ := codecMap[pack.ChunkId]
			if decoder.done{
				fmt.Printf("[Chunk %d]: Chunk %d already finished...\n",pack.ChunkId,pack.ChunkId)
				go SendFin(pack.ChunkId,conn,meta.FileName)
				continue
			}
			result := decoder.decoder.Decode([]codec.LTBlock{{
				BlockCode: pack.BlockId,
				Data:      rawData,
			}})
			if result == codec.DECODE_NEEDMORE {
				fmt.Printf("[Chunk %d]: Package %d received, need more...\n",pack.ChunkId,pack.BlockId)
				continue
			} else {
				fmt.Printf("[Chunk %d]: Chunk %d finished!\n",pack.ChunkId,pack.ChunkId)
				decoder.done = true
				codecMap[pack.ChunkId] = decoder
				writeFileWg.Add(1)
				go func() {
					// write to cache
					defer writeFileWg.Done()
					recoverd, err := decoder.decoder.Recover()
					if err!=nil{
						fmt.Printf("Decode chunk %d failed.\n",pack.ChunkId)
						println(err.Error())
						decoder.done = false
						codecMap[pack.ChunkId] = decoder
						return
					}
					err = utils.WriteToCache(recoverd,meta.FileName, int(pack.ChunkId), int(pack.Length))
					if err!=nil{
						fmt.Printf("Write chunk %d failed.\n",pack.ChunkId)
						println(err.Error())
						decoder.done = false
						codecMap[pack.ChunkId] = decoder
						return
					}
					SendFin(pack.ChunkId,conn,meta.FileName)
					if pack.Length < meta.PackageSize{
						println("[Info]: All chunks finished..")
						isFin = true
					}
				}()
			}
		}
	}
	writeFileWg.Wait()
	return utils.RecoveryFromCache(meta.FileName,saveTo)
}

func SendFin(chunkId int64, conn net.Conn, filename string) {
	finMsg := protocol.Fin{
		FileName: filename,
		ChunkId:  chunkId,
		Done:     true,
	}
	finStr, _ := finMsg.Marshal()
	_, _ = conn.Write(finStr)
}