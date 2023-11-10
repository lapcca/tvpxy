package proxy

import (
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net"
	"os"
	"time"
)

type Proxy struct {
	inteface *net.Interface
	timeout  time.Duration
	engine   *gin.Engine
	logger   *log.Logger
}

func NewProxy(netDeviceName string, timeout string) *Proxy {
	engine := gin.Default()

	p := &Proxy{}
	p.engine = engine
	p.parseAndSetInterface(netDeviceName)
	p.setTimeout(timeout)
	p.setLogger()
	return p
}

func (p *Proxy) setLogger() {

	logger := log.New(io.MultiWriter(os.Stdout), "", log.Lshortfile|log.LstdFlags)
	p.logger = logger
}

func (p *Proxy) Run(port string) {
	p.engine.GET("/rtp/:udp_addr", p.handle)

	if port == "" {
		port = "5555"
	}
	portStr := ":" + port
	p.logger.Println("starting server on " + portStr)
	err := p.engine.Run(portStr)
	if err != nil {
		p.logger.Fatalf("error starting server: %v", err)
	}
}

func (p *Proxy) parseAndSetInterface(netDeviceName string) {

	inf, err := net.InterfaceByName(netDeviceName)
	if err != nil {
		p.logger.Printf("error parsing interface: %v", err)
	}
	p.inteface = inf
}

func (p *Proxy) setTimeout(timeout string) {
	timeoutDur, err := time.ParseDuration(timeout)
	if err != nil {
		p.logger.Printf("error parsing duration: %v", err)
	}
	p.timeout = timeoutDur
}

func (p *Proxy) handle(c *gin.Context) {
	udpAddr := c.Param("udp_addr")
	addr, err := net.ResolveUDPAddr("udp4", udpAddr)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	p.logger.Printf("got request for [%s]\n", udpAddr)

	conn, err := net.ListenMulticastUDP("udp4", p.inteface, addr)
	if err != nil {
		c.String(500, err.Error())
		return
	}
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		if err != nil {
			log.Fatalf("error closing connection: %v", err)
		}
	}(conn)

	err = conn.SetReadDeadline(time.Now().Add(p.timeout))
	if err != nil {
		c.String(500, err.Error())
		return
	}

	var buf = make([]byte, 1500)
	num, err := conn.Read(buf)
	if err != nil {
		c.String(500, err.Error())
		return
	}

	err = conn.SetReadDeadline(time.Time{})
	if err != nil {
		c.String(500, err.Error())
		return
	}

	headerSent := false
	for {

		payloadType := buf[1] & 0x7F
		playLoad := buf[12:num]

		if !headerSent {
			headerSent = true
			if payloadType == RTP_Payload_MP2T {
				c.Writer.Header().Set("Content-Type", ContentType_MP2T)
			} else {
				c.Writer.Header().Set("Content-Type", ContentType_DEFAULT)
			}
			c.Writer.WriteHeader(200)
		}

		if _, wErr := c.Writer.Write(playLoad); wErr != nil {
			break
		}

		if num, err = conn.Read(buf); err != nil {
			break
		}
	}
	if err != nil && err != io.EOF {
		c.String(500, err.Error())
		return
	}
}
