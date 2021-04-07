package utils

import (
	"bufio"
	"fmt"
	"github.com/Gaojianli/well_of_file/config"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sort"
	"strconv"
)

func WriteToCache(buffer []byte, name string, id, length int) error {
	cachePath := path.Join(config.TEMP_PATH, config.CACHE_DIR)
	err := os.MkdirAll(cachePath, os.ModePerm)
	if err != nil {
		panic(err.Error())
	}
	chunkPath := fmt.Sprintf("%s.%d", path.Join(cachePath, name), id)
	fs, err := os.OpenFile(chunkPath, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer fs.Close()
	_, err = fs.Write(buffer[:length])
	if err != nil {
		return err
	}
	fmt.Printf("[Chunk %d]: Chunk write to %s\n", id, chunkPath)
	return err
}

func RecoveryFromCache(name, saveTo string) error {
	fs, err := os.OpenFile(path.Join(saveTo, name), os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer fs.Close()
	cachePath := path.Join(config.TEMP_PATH, config.CACHE_DIR)
	fileInfos, err := ioutil.ReadDir(cachePath)
	if err != nil {
		println("Failed to read cache dir\n")
		log.Fatal(err.Error())
	}
	sort.Slice(fileInfos, func(i, j int) bool {
		fileName1 := fileInfos[i].Name()
		fileName2 := fileInfos[j].Name()
		fileNum1, _ := strconv.ParseInt(fileName1[len(name)+1:], 10, 32)
		fileNum2, _ := strconv.ParseInt(fileName2[len(name)+1:], 10, 32)
		return fileNum1 < fileNum2
	})
	for _, fileinfo := range fileInfos {
		eof := false
		chunkFs, err := os.Open(path.Join(cachePath, fileinfo.Name()))
		if err != nil {
			return err
		}
		r := bufio.NewReader(chunkFs)
		fileBuffer := make([]byte, config.CHUNK_SZIE)
		for !eof {
			count, err := r.Read(fileBuffer)
			if err == io.EOF {
				eof = true
			} else if err != nil {
				return err
			}
			_, err = fs.Write(fileBuffer[:count])
			if err != nil {
				return err
			}
		}
	}
	fmt.Printf("[Info]: File write to %s\n", path.Join(saveTo, name))
	return nil
}
