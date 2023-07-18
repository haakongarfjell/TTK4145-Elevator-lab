package localip

import (
	"net"
	"strings"
)

var local_IP string

func LocalIP() (string, error) {
	if local_IP == "" {
		conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
		if err != nil {
			return "", err
		}
		defer conn.Close()
		local_IP = strings.Split(conn.LocalAddr().String(), ":")[0]
	}
	return local_IP, nil
}
