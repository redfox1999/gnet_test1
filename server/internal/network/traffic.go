package network

import (
	"encoding/binary"
	"log"

	"gnet_test1/internal/protocol"

	"github.com/panjf2000/gnet/v2"
)

func (gs *GatewayServer) OnTraffic(c gnet.Conn) gnet.Action {
	var finalAction gnet.Action = gnet.None

	for {
		inboundLen := c.InboundBuffered()
		if inboundLen < protocol.HeaderLen {
			break
		}

		header, _ := c.Peek(protocol.HeaderLen)
		cmdID := binary.BigEndian.Uint32(header[0:4])
		dataLen := binary.BigEndian.Uint32(header[4:8])

		totalLen := int(protocol.HeaderLen) + int(dataLen)
		if totalLen > gs.cfg.MaxPacketSize || dataLen < 0 {
			log.Printf("[安全警告] 收到非法长度的恶意包: dataLen=%d，强制掐断客户端: %s", dataLen, c.RemoteAddr().String())
			return gnet.Close
		}

		if inboundLen < totalLen {
			break
		}

		fullPacket, err := c.Next(totalLen)
		if err != nil {
			return gnet.Close
		}

		payload := fullPacket[int(protocol.HeaderLen):totalLen]
		payloadCopy := make([]byte, len(payload))
		copy(payloadCopy, payload)

		action := gs.dispatchBusiness(c, cmdID, payloadCopy)
		if action != gnet.None {
			finalAction = action
			break
		}
	}

	return finalAction
}
