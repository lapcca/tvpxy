package proxy

import (
	"github.com/gin-gonic/gin"
	"github.com/pion/rtp"
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
	p.setLogger("")
	return p
}

func (p *Proxy) setLogger(logFileName string) {

	if logFileName == "" {
		logFileName = "log.txt"
	}
	logHandler, err := os.OpenFile(logFileName, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		log.Fatalf("create file %v failed: %v", logFileName, err)
	}

	logger := log.New(io.MultiWriter(os.Stdout, logHandler), "Proxy:", log.Lshortfile|log.LstdFlags)
	p.logger = logger
}

func (p *Proxy) Run(port string) {
	p.engine.GET("/rtp/:udp_addr", p.handle)

	if port == "" {
		port = "5566"
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
		p.logger.Printf("%v:error resolving address: %v", c.Param("udp_addr"), udpAddr)
		return
	}
	p.logger.Printf("Got request for %s\n", c.Param("udp_addr"))

	conn, err := net.ListenMulticastUDP("udp4", p.inteface, addr)
	if err != nil {
		c.String(500, err.Error())
		p.logger.Printf("%v: error listening: %v", c.Param("udp_addr"), err.Error())
		return
	}
	defer func(conn *net.UDPConn) {
		err := conn.Close()
		if err != nil {
			p.logger.Printf("%v:error closing connection: %v", c.Param("udp_addr"), err)
		}
	}(conn)

	err = conn.SetReadDeadline(time.Now().Add(p.timeout))
	if err != nil {
		c.String(500, err.Error())
		p.logger.Printf("%v:error setting deadline: %v", c.Param("udp_addr"), err.Error())
		return
	}

	channel := make(chan DATA, 700)
	endChannel := make(chan bool)
	go p.receiveRoutine(channel, endChannel, conn, c)
	go p.sendRoutine(channel, endChannel, c)

	<-endChannel

}

func (p *Proxy) sendRoutine(channel chan DATA, endChannel chan bool, c *gin.Context) {

	var err error
	var data DATA
	var wNum int

	rtpPkg := &rtp.Packet{}
	headerSent := false
	for {
		select {
		case data = <-channel:
			num := data.len
			if err = rtpPkg.Unmarshal(data.buf[:num]); err != nil {
				c.String(500, err.Error())
				p.logger.Printf("%v:error unmarshalling: %v", c.Param("udp_addr"), err.Error())
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

			if wNum, err = c.Writer.Write(rtpPkg.Payload); err != nil {
				c.String(500, err.Error())
				endChannel <- true
				p.logger.Printf("%v:error writing: %v", c.Param("udp_addr"), err.Error())
				return
			}
			if wNum != num {
				p.logger.Printf("Just write %d of %d bytest data", wNum, num)
			}

		case <-endChannel:
			return
		}

	}
}

func (p *Proxy) receiveRoutine(channel chan DATA, endChannel chan bool, conn *net.UDPConn, c *gin.Context) {
	var buf = make([]byte, 1500)

	num, err := conn.Read(buf)
	if err != nil {
		c.String(500, err.Error())
		p.logger.Printf("%v:error reading: %v", c.Param("udp_addr"), err.Error())
		return
	}
	p.logger.Printf("Got %d bytes from %s\n", num, c.Param("udp_addr"))

	err = conn.SetReadDeadline(time.Time{})
	if err != nil {
		c.String(500, err.Error())
		p.logger.Printf("%v:error setting deadline: %v", c.Param("udp_addr"), err.Error())
		return
	}

	for {

		select {
		case channel <- DATA{
			buf: buf[:num],
			len: num,
		}:

		case <-endChannel:
			return
		}
		buf = []byte{}

		if num, err = conn.Read(buf); err != nil {
			p.logger.Printf("%v:error reading: %v", c.Param("udp_addr"), err.Error())
			break
		}
	}

	if err != nil && err != io.EOF {
		c.String(500, err.Error())
		endChannel <- true
		p.logger.Printf("%v:error reading last: %v", c.Param("udp_addr"), err.Error())
		return
	}
}
