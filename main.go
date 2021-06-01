package main

//package main
import (
	"flag"
	"os"
)

func main() {
	isServer := flag.Bool("server", false, "Running mode,can be server,client")
	port := flag.Int("p", 11451, "Port")
	input := flag.String("i", "", "File input")
	address := flag.String("address", "", "[Client]Server Address")
	output := flag.String("o", "./", "[Client]File output")
	flag.Parse()
	if *isServer {
		if len(*input) == 0 {
			flag.Usage()
			os.Exit(-1)
		}
		Send(*input, *port)
	} else {
		SetPath(*output, "lt_cache")
		if len(*address) == 0 {
			flag.Usage()
			os.Exit(-1)
		}
		Receive(*address, *port, *output)
	}
}
