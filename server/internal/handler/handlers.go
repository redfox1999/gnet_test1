package handler

import (
	"encoding/binary"

	"gnet_test1/internal/protocol"

	"github.com/panjf2000/gnet/v2"
)

// CalculateHandler 处理计算请求
type CalculateHandler struct{}

func (h *CalculateHandler) Handle(c gnet.Conn, cmdID uint16, body []byte) {
	if len(body) < 8 {
		return
	}
	a := binary.BigEndian.Uint32(body[0:4])
	b := binary.BigEndian.Uint32(body[4:8])
	cResult := a + b
	result := make([]byte, 12)
	binary.BigEndian.PutUint32(result[0:4], a)
	binary.BigEndian.PutUint32(result[4:8], b)
	binary.BigEndian.PutUint32(result[8:12], cResult)
	protocol.SendPacket(c, uint32(cmdID), result)
}

// SmallHandler 处理小数据包请求
type SmallHandler struct{}

func (h *SmallHandler) Handle(c gnet.Conn, cmdID uint16, body []byte) {
	//response := "small response"
	protocol.SendPacket(c, uint32(cmdID), body)
}

// MediumHandler 处理中等数据包请求
type MediumHandler struct{}

func (h *MediumHandler) Handle(c gnet.Conn, cmdID uint16, body []byte) {
	//response := "medium response"
	protocol.SendPacket(c, uint32(cmdID), body)
}

// LargeHandler 处理大数据包请求
type LargeHandler struct{}

func (h *LargeHandler) Handle(c gnet.Conn, cmdID uint16, body []byte) {
	//response := "large response"
	protocol.SendPacket(c, uint32(cmdID), body)
}
