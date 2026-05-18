package network

import (
	"encoding/binary"
	"log"

	"github.com/panjf2000/gnet/v2"
)

func (gs *GatewayServer) OnTraffic(c gnet.Conn) gnet.Action {
	var finalAction gnet.Action = gnet.None

	for {
		inboundLen := c.InboundBuffered()
		if inboundLen < HeaderLen {
			break
		}

		sizeBuf, _ := c.Peek(4)
		totalLen := int(binary.BigEndian.Uint32(sizeBuf))

		if totalLen > gs.cfg.MaxPacketSize || totalLen < HeaderLen {
			log.Printf("[安全警告] 收到非法长度的恶意包: %d 字节，强制掐断客户端: %s", totalLen, c.RemoteAddr().String())
			return gnet.Close
		}

		if inboundLen < totalLen {
			break
		}

		fullPacket, err := c.Next(totalLen)
		if err != nil {
			return gnet.Close
		}

		cmdID := binary.BigEndian.Uint16(fullPacket[4:6])
		seqNum := binary.BigEndian.Uint32(fullPacket[6:10])
		payload := fullPacket[HeaderLen:totalLen]

		payloadCopy := make([]byte, len(payload))
		copy(payloadCopy, payload)

		action := gs.dispatchBusiness(c, cmdID, seqNum, payloadCopy)
		if action != gnet.None {
			finalAction = action
			break
		}
	}

	return finalAction
}
