package lib

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
)

var LocalIP = GetLocalIPs()[0]

// NewTrace 创建新的 TraceContext 对象，并生成唯一的 TraceId 和 SpanId
// 分布式链路追踪：通过生成唯一的 TraceId 和 SpanId，
// 可以在监控系统或日志系统中跟踪每个请求在整个系统中的路径，识别请求是如何流经各个服务的。
func NewTrace() *TraceContext {
	trace := &TraceContext{}
	trace.TraceId = GetTraceId() //唯一标识
	trace.SpanId = NewSpanId()
	return trace
}

// NewSpanId 生成唯一的 SpanId。
// 通过 XOR 操作结合 IP 地址的整数值和当前时间戳，再加上一个随机数，构造出一个唯一的 SpanId。
// 这可以帮助精细化追踪每个请求操作中的步骤，特别是在分布式环境中非常有用。
func NewSpanId() string {
	timestamp := uint32(time.Now().Unix())

	ipToLong := binary.BigEndian.Uint32(LocalIP.To4())
	b := bytes.Buffer{}
	b.WriteString(fmt.Sprintf("%08x", ipToLong^timestamp))
	b.WriteString(fmt.Sprintf("%08x", rand.Int31()))
	return b.String()
}

// GetTraceId 基于本地 IP 地址、时间戳等信息生成的，确保每个请求有唯一的标识。
func GetTraceId() (traceId string) {
	return calcTraceId(LocalIP.String())
}

// calcTraceId 函数中，结合了时间戳、进程 ID (pid) 以及随机数，再加上 IP 地址，
// 生成了一个具有高唯一性的追踪 ID，确保同一台机器不同请求的追踪 ID 也是唯一的。
func calcTraceId(ip string) (traceId string) {
	now := time.Now()
	timestamp := uint32(now.Unix())
	timeNano := now.UnixNano()
	pid := os.Getpid()

	b := bytes.Buffer{}
	netIP := net.ParseIP(ip)
	if netIP == nil {
		b.WriteString("00000000")
	} else {
		b.WriteString(hex.EncodeToString(netIP.To4()))
	}
	b.WriteString(fmt.Sprintf("%08x", timestamp&0xffffffff))
	b.WriteString(fmt.Sprintf("%04x", timeNano&0xffff))
	b.WriteString(fmt.Sprintf("%04x", pid&0xffff))
	b.WriteString(fmt.Sprintf("%06x", rand.Int31n(1<<24)))
	b.WriteString("b0") // 末两位标记来源,b0为go

	return b.String()
}

func GetLocalIPs() (ips []net.IP) {
	interfaceAddr, err := net.InterfaceAddrs()
	if err != nil {
		return nil
	}
	for _, address := range interfaceAddr {
		ipNet, isValidIpNet := address.(*net.IPNet)
		if isValidIpNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP)
			}
		}
	}
	//fmt.Println(ips)
	return ips
}
