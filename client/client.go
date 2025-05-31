package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"

	"golang.org/x/net/ipv4"

	"github.com/Bokai-Bi/Relay/pkg/relay"
	"github.com/songgao/water"
)

func main() {
	testKey := []byte{129,50,227,239,129,50,227,239,129,50,227,239,129,50,227,239,}
	testIFIP := "192.168.5.14/24"
	testServerIP := "129.213.102.19"
	iface, err := createTun(testIFIP)
	if err != nil {
		fmt.Println(err)
	}
	stopChan := make(chan bool)
	relayClient := relay.MakeRelayClient(testServerIP, testKey)
	go listenInterface(relayClient, iface, stopChan)
	go listenInbound(relayClient, iface, stopChan)
	// Wait for user to input "stop"
	var input string
	fmt.Println("Press Enter to stop...")
	fmt.Scanln(&input)
	stopChan <- true
	iface.Close()
}

func RunCommand(command string) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("bash", "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if stderr.String() != "" {
		return stderr.String(), err
	}
	return stdout.String(), err
}

func createTun(ip string) (*water.Interface, error) {
	config := water.Config{
		DeviceType: water.TUN,
	}

	iface, err := water.New(config)
	if err != nil {
		return nil, err
	}
	log.Printf("Interface Name: %s\n", iface.Name())

	out, err := RunCommand(fmt.Sprintf("sudo ip link set up dev %s", iface.Name()))
	if err != nil {
		fmt.Println(out)
		return nil, err
	}

	out, err = RunCommand(fmt.Sprintf("sudo ip addr add %s dev %s", ip, iface.Name()))
	if err != nil {
		fmt.Println(out)
	}

	
	return iface, nil
}

func listenInterface(client *relay.RelayClient, iface *water.Interface, stopChan chan bool) {
	packet := make([]byte, 65535)
	for {
		n, err := iface.Read(packet)
		if err != nil {
			log.Println("ifce read error:", err)
		}
		if err == nil {
			fmt.Printf("Received packet size %d content %d\n", n, packet[:n])
			PrintPacket((packet[:n]))
			client.ForwardPacket(packet[:n])
		}
	}
}


func PrintPacket(frame []byte) {
	header, err := ipv4.ParseHeader(frame)
	if err != nil {
		fmt.Println("write packet err:", err)
	} else {
		fmt.Println("SRC:", header.Src)
		fmt.Println("DST:", header.Dst)
		fmt.Println("ID:", header.ID)
		fmt.Println("CHECKSUM:", header.Checksum)
	}
}

func listenInbound(client *relay.RelayClient, iface *water.Interface, stopChan chan bool) {
	packet := make([]byte, 65535)
	for {
		packetLen := client.ReceivePacket(packet)
		if packetLen == 0 {
			log.Println("Error receiving packet or packet size is zero")
			continue
		}
		bytesWritten := 0
		for bytesWritten < len(packet) {
			
	}
}