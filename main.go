package main

import (
	"flag"
	"fmt"
	"github.com/lapcca/tvpxy/proxy"
	"io"
	"log"
	"os"
)

func main() {

	srcNet := flag.String("net", "eth4", "the network device for RTP packets")
	port := flag.String("port", "5566", "the port of the server run on")
	timeout := flag.String("timeout", "30s", "The timeout for waiting for RTP packets")

	flag.Parse()
	logger := NewLogger("")
	logger.Println("start proxy")
	proxyServer := proxy.NewProxy(*srcNet, *timeout)
	proxyServer.SetLogger(logger)
	proxyServer.Run(*port)
}

func NewLogger(logFileName string) *log.Logger {
	fmt.Println("set logger")
	if logFileName == "" {
		logFileName = "log.txt"
	}
	logHandler, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		fmt.Println("error opening log file")
		log.Fatalf("create file %vfailed: %v", logFileName, err)
	}

	logger := log.New(io.MultiWriter(os.Stdout, logHandler), "Proxy:", log.Lshortfile|log.LstdFlags)
	return logger
}
