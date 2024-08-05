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
	testKey := []byte{129,50,227,239}
	testIFIP := "192.168.9.9"
	testServerIP := "129.213.102.19"
	iface, err := createTun(testIFIP)
	if err != nil {
		fmt.Println(err)
	}
	stopChan := make(chan bool)
	relayClient := relay.MakeRelayClient(testServerIP, testKey)
	go listenInterface(relayClient, iface, stopChan)
	go listenInbound(relayClient, iface, stopChan)
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
	out, err := RunCommand(fmt.Sprintf("sudo ip addr add %s/24 dev %s", ip, iface.Name()))
	if err != nil {
		fmt.Println(out)
	}

	out, err = RunCommand(fmt.Sprintf("sudo ip link set dev %s up", iface.Name()))
	if err != nil {
		fmt.Println(out)
		return nil, err
	}
	return iface, nil
}

func listenInterface(client *relay.RelayClient, iface *water.Interface, stopChan chan bool) {
	fmt.Println("interface listening")
	packet := make([]byte, 65535)
	for {
		n, err := iface.Read(packet)
		if err != nil {
			log.Println("ifce read error:", err)
		}
		if err == nil {
			WritePacket((packet[:n]))
			
		}
	}
}

func WritePacket(frame []byte) {
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