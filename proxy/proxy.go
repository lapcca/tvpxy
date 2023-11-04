package proxy

import (
	"github.com/gin-gonic/gin"
	"github.com/pion/rtp"
	"io"
	"log"
	"net"
	"time"
)

type Proxy struct {
	inteface *net.Interface
	timeout  time.Duration
	engine   *gin.Engine
}

func NewProxy(netDeviceName string, timeout string) *Proxy {
	engine := gin.Default()

	p := &Proxy{}
	p.engine = engine
	p.parseAndSetInterface(netDeviceName)
	p.setTimeout(timeout)

	return p
}

func (p *Proxy) Run(port string) {
	p.engine.GET("/rtp/:udp_addr", p.handle)

	if port == "" {
		port = "5566"
	}
	portStr := ":" + port
	err := p.engine.Run(portStr)
	if err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}

func (p *Proxy) parseAndSetInterface(netDeviceName string) {

	inf, err := net.InterfaceByName(netDeviceName)
	if err != nil {
		log.Fatalf("error parsing interface: %v", err)
	}
	p.inteface = inf
}

func (p *Proxy) setTimeout(timeout string) {
	timeoutDur, err := time.ParseDuration(timeout)
	if err != nil {
		log.Fatalf("error parsing duration: %v", err)
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

	rtpPkg := &rtp.Packet{}
	headerSent := false
	for {
		if err = rtpPkg.Unmarshal(buf[:num]); err != nil {
			c.String(500, err.Error())
			return
		}

		if !headerSent {
			headerSent = true
			if rtpPkg.PayloadType == RTP_Payload_MP2T {
				c.Writer.Header().Set("Content-Type", ContentType_MP2T)
			} else {
				c.Writer.Header().Set("Content-Type", ContentType_DEFAULT)
			}
			c.Writer.WriteHeader(200)
		}

		if _, wErr := c.Writer.Write(rtpPkg.Payload); wErr != nil {
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
