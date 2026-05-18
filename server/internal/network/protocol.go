package network

import (
	"bytes"
	"encoding/binary"
	"log"

	"github.com/panjf2000/gnet/v2"
)

const (
	HeaderLen = 10

	CmdLogin    uint16 = 1001
	CmdChat     uint16 = 1002
	CmdShutDown uint16 = 9999
)

func SendPacket(c gnet.Conn, cmdID uint16, seq uint32, body []byte) {
	bodyLen := len(body)
	totalLen := HeaderLen + bodyLen

	buf := bytes.NewBuffer(make([]byte, 0, totalLen))

	_ = binary.Write(buf, binary.BigEndian, uint32(totalLen))
	_ = binary.Write(buf, binary.BigEndian, cmdID)
	_ = binary.Write(buf, binary.BigEndian, seq)
	_, _ = buf.Write(body)

	err := c.AsyncWrite(buf.Bytes(), nil)
	if err != nil {
		log.Printf("[发送失败] 无法投递回包给 %s: %v", c.RemoteAddr().String(), err)
	}
}
