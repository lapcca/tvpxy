package main

import (
	"flag"
	"github.com/lapcca/tvpxy/proxy"
)

func main() {

	srcNet := flag.String("net", "eth4", "the network device for RTP packets")
	port := flag.String("port", "5566", "the port of the server run on")
	timeout := flag.String("timeout", "30s", "The timeout for waiting for RTP packets")

	flag.Parse()
	proxyServer := proxy.NewProxy(*srcNet, *timeout)
	proxyServer.Run(*port)
}
