package protocol

import (
	"bytes"
	"encoding/binary"
	"log"
	"net"

	"github.com/panjf2000/gnet/v2"
)

const (
	HeaderLen = 8

	CmdCalculate uint32 = 1001
	CmdSmall     uint32 = 2001
	CmdMedium    uint32 = 3001
	CmdLarge     uint32 = 4001
	CmdShutDown  uint32 = 9999
)

// Conn 定义连接接口
type Conn interface {
	RemoteAddr() net.Addr
	AsyncWrite([]byte, gnet.AsyncCallback) error
}

func SendPacket(c Conn, cmdID uint32, body []byte) {
	dataLen := len(body)
	totalLen := HeaderLen + dataLen

	buf := bytes.NewBuffer(make([]byte, 0, totalLen))

	_ = binary.Write(buf, binary.BigEndian, cmdID)
	_ = binary.Write(buf, binary.BigEndian, uint32(dataLen))
	_, _ = buf.Write(body)

	err := c.AsyncWrite(buf.Bytes(), nil)
	if err != nil {
		log.Printf("[发送失败] 无法投递回包给 %s: %v", c.RemoteAddr().String(), err)
	}
}
