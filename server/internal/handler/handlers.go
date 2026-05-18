package handler

import (
	"gnet_test1/internal/protocol"
	"log"

	"github.com/panjf2000/gnet/v2"
)

// CalculateHandler 处理计算请求
type CalculateHandler struct{}

func (h *CalculateHandler) Handle(c gnet.Conn, cmdID uint16, body []byte) {
	log.Printf("[业务-Calculate] 收到计算请求: %s", string(body))
	result := "calculated result"
	protocol.SendPacket(c, uint32(cmdID), []byte(result))
}

// SmallHandler 处理小数据包请求
type SmallHandler struct{}

func (h *SmallHandler) Handle(c gnet.Conn, cmdID uint16, body []byte) {
	log.Printf("[业务-Small] 收到小数据包: %s", string(body))
	response := "small response"
	protocol.SendPacket(c, uint32(cmdID), []byte(response))
}

// MediumHandler 处理中等数据包请求
type MediumHandler struct{}

func (h *MediumHandler) Handle(c gnet.Conn, cmdID uint16, body []byte) {
	log.Printf("[业务-Medium] 收到中等数据包，长度: %d", len(body))
	response := "medium response"
	protocol.SendPacket(c, uint32(cmdID), []byte(response))
}

// LargeHandler 处理大数据包请求
type LargeHandler struct{}

func (h *LargeHandler) Handle(c gnet.Conn, cmdID uint16, body []byte) {
	log.Printf("[业务-Large] 收到大数据包，长度: %d", len(body))
	response := "large response"
	protocol.SendPacket(c, uint32(cmdID), []byte(response))
}
