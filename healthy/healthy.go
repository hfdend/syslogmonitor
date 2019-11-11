package healthy

import (
	"net"
	"time"
)

func TcpCheck(addr, port string, timeout time.Duration) bool {
	var (
		remote = addr + ":" + port
	)
	conn, err := net.DialTimeout("tcp", remote, timeout)
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}
