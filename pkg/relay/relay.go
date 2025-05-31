package relay

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/google/netstack/tcpip/header"
)

type RelayClient struct {
	serverIP *net.TCPAddr
	serverConn *net.TCPConn
	// encryptor *relaycrypt.AESEncryptor
}

const MAX_IP_DATA_SIZE int = 65515
func MakeRelayClient(server string, encryptKey []byte) *RelayClient {
	client := new(RelayClient)
	serverIP, err := net.ResolveTCPAddr("tcp4", server)
	if err != nil {
		fmt.Errorf("Cannot parse server to ip and port, server: ", server)
	}
	client.serverIP = serverIP
	client.serverConn, err = net.DialTCP("tcp4", nil, serverIP)
	if err != nil {
		fmt.Errorf("Cannot dial to server, ", serverIP)
	}
	// client.encryptor = relaycrypt.MakeAES128Encryptor(encryptKey)
	
	return client
}

func (client *RelayClient) ForwardPacket(packet []byte) {
	bytesWritten := 0
	for bytesWritten < len(packet) {
		written, err := client.serverConn.Write(packet)
		if err != nil {
			fmt.Println("Error writing to server: ", err)
			return
		}
		bytesWritten += written
	}
}

func (client *RelayClient) ReceivePacket(buffer []byte) int {
	bytesRead := 0
	for bytesRead < 4 {
		n, err := client.serverConn.Read(buffer[bytesRead:4])
		if err != nil {
			fmt.Println("Error reading from server: ", err)
			return 0
		}
		bytesRead += n
	}
	packetSize := int(binary.BigEndian.Uint16(buffer[2:4]))
	for bytesRead < packetSize {
		n, err := client.serverConn.Read(buffer[bytesRead:packetSize])
		if err != nil {
			fmt.Println("Error reading from server: ", err)
			return 0
		}
		bytesRead += n
	}
	return packetSize
}

/* // Wrap the content of an ip packet inside a relay packet and return the packet to send
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
} */

/* func ReliableWrite(conn net.Conn, data []byte, size int, protocol int) error {
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
} */

type RelayServer struct {
	ListeningPort int
	ListeningConn *net.TCPListener
}

func MakeRelayServer(port int) *RelayServer {
	server := new(RelayServer)
	serverIP, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		fmt.Errorf("Cannot parse server to ip and port, port: %d", port)
		return nil
	}
	server.ListeningPort = port
	server.ListeningConn, err = net.ListenTCP("tcp4", serverIP)
	if err != nil {
		fmt.Errorf("Cannot listen to port %d, error: %v", port, err)
		return nil
	}
	return server
}

func (server *RelayServer) AcceptConnection() (*net.TCPConn, error) {
	conn, err := server.ListeningConn.AcceptTCP()
	if err != nil {
		fmt.Errorf("Error accepting connection: %v", err)
		return nil, err
	}
	fmt.Printf("Accepted connection from %s\n", conn.RemoteAddr().String())
	return conn, nil
}

func (server *RelayServer) HandleClient(conn *net.TCPConn) {
	defer conn.Close()
	buffer := make([]byte, 65535)
	for {
		packetSize := server.ReceivePacket(conn, buffer)
		if packetSize == 0 {
			fmt.Println("No data received or error occurred")
			return
		}
		fmt.Printf("Received packet of size %d\n", packetSize)
		
		// extract source port, destination ip and port
		// parse as ip packet
		ipHeader := header.IPv4(buffer[:packetSize])
		if !ipHeader.IsValid(packetSize) {
			fmt.Println("Invalid IP header")
			return
		}

		dstIP := ipHeader.DestinationAddress()
		dstPort := binary.BigEndian.Uint16(ipHeader.Payload()[2:4])
		srcPort := binary.BigEndian.Uint16(ipHeader.Payload()[0:2])
	}
}

func (server *RelayServer) ReceivePacket(conn *net.TCPConn, buffer []byte) int {
	bytesRead := 0
	for bytesRead < 4 {
		n, err := conn.Read(buffer[bytesRead:4])
		if err != nil {
			fmt.Println("Error reading from client: ", err)
			return 0
		}
		bytesRead += n
	}
	packetSize := int(binary.BigEndian.Uint16(buffer[2:4]))
	for bytesRead < packetSize {
		n, err := conn.Read(buffer[bytesRead:packetSize])
		if err != nil {
			fmt.Println("Error reading from client: ", err)
			return 0
		}
		bytesRead += n
	}
	return packetSize
}


func ComputeChecksum(b []byte) uint16 {
	checksum := header.Checksum(b, 0)

	checksumInv := checksum ^ 0xffff

	return checksumInv
}