package relay

import (
	"fmt"
	"net"
	"net/netip"
	"netip"

	"github.com/google/netstack/tcpip/header"
	"golang.org/x/net/ipv4"

	"github.com/Bokai-Bi/Relay/internal/relaycrypt"
)

type RelayClient struct {
	serverIP net.IP
	serverConn net.Conn
	encryptor *relaycrypt.AESEncryptor
	recvBuffer []byte // buffer used to get incoming data
	sendBuffer []byte // buffer used to store data before sending
	sendFragmentBuffer []byte // in case the additional encrypted target ip exceeds max ip packet size, fragment
}

const MAX_IP_DATA_SIZE int = 65515
func MakeRelayClient(server string, encryptKey []byte) *RelayClient {
	client := new(RelayClient)
	serverIP := net.ParseIP(server)
	if serverIP == nil {
		fmt.Errorf("Cannot parse server to ip, server: ", server)
	}
	client.serverIP = serverIP
	temp, err := net.Dial("ip4", server)
	if err != nil {
		fmt.Errorf("Cannot dial to server, ", server)
	}
	client.serverConn = temp
	client.encryptor = relaycrypt.MakeAES128Encryptor(encryptKey)
	client.recvBuffer = make([]byte, MAX_IP_DATA_SIZE)
	client.sendBuffer = make([]byte, MAX_IP_DATA_SIZE)
	client.sendFragmentBuffer = make([]byte, MAX_IP_DATA_SIZE)
	return client
}

// Wrap the content of an ip packet inside a relay packet and return the packet to send
func (client *RelayClient) SendRelayPacket(data []byte) error {
	header, err := ipv4.ParseHeader(data)
	if err != nil {
		fmt.Println("Failed to parse header ", err)
		return err
	}
	protocol := header.Protocol
	trueDst := header.Dst
	copy(trueDst[0:], client.encryptor.NextNonce)
	encSize := len(client.encryptor.AES128EncryptIP(trueDst[relaycrypt.NonceSize:], client.sendBuffer)) + relaycrypt.NonceSize
	if (header.TotalLen - header.Len > MAX_IP_DATA_SIZE - encSize) {
		midPoint := (header.TotalLen - header.Len) / 2
		copy(client.sendBuffer[encSize:], data[header.Len:header.Len + midPoint])
		ReliableWrite(client.serverConn, client.sendBuffer, encSize + midPoint, protocol)
		copy(client.sendBuffer[encSize:], data[header.Len+midPoint : header.TotalLen])
		ReliableWrite(client.serverConn, client.sendBuffer, encSize + header.TotalLen - header.Len - midPoint, protocol)
	} else {
		copy(client.sendBuffer[encSize:], data[header.Len:header.TotalLen])
		sz := encSize + header.TotalLen - header.Len
		ReliableWrite(client.serverConn, client.sendBuffer, sz, protocol)
	}
	return nil
}

func ReliableWrite(conn net.Conn, data []byte, size int, protocol int) error {
	written := 0
	for (written < size) {
		w, err := conn.Write(data[written:size])
		if err != nil {
			fmt.Errorf("Error when writing: ", err)
			return err
		}
		written += w
	}
	return nil
}

type RelayServer struct {
	encryptKey []byte
	forwardList map[netip.AddrPort] netip.AddrPort
}
// Unwrap the content of 
func forwardRelayPacket() {

}

func ComputeChecksum(b []byte) uint16 {
	checksum := header.Checksum(b, 0)

	checksumInv := checksum ^ 0xffff

	return checksumInv
}