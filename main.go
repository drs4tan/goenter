package entergo

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/pion/dtls/v2"
	"github.com/pion/dtls/v2/examples/util"
	"net"
	"net/http"
	"time"
)

type ApiActive struct {
	Active bool `json:"active"`
}

type Apireq struct {
	Stream ApiActive `json:"stream"`
}

type LightPacket struct {
	LightID uint8
	Red     uint8
	Green   uint8
	Blue    uint8
}

// BeginStream starts the intitial connection
func BeginStream(hubip string, authcode string, lightgroup string, clientcode string) (dtlsconn *dtls.Conn, channel chan []LightPacket, cancel func()) {
	start := time.Now()
	startsession(hubip, authcode, lightgroup)
	duration := time.Since(start)
	fmt.Printf("Server started session in %v\n", duration)
	// Prepare the IP to connect to
	addr := &net.UDPAddr{IP: net.ParseIP(hubip), Port: 2100}

	// Prepare the configuration of the DTLS connection
	config := &dtls.Config{
		PSK: func(hint []byte) ([]byte, error) {
			test, _ := hex.DecodeString(clientcode)
			return test, nil
		},
		PSKIdentityHint: []byte(authcode),
		CipherSuites:    []dtls.CipherSuiteID{dtls.TLS_PSK_WITH_AES_128_GCM_SHA256},
	}

	// Connect to a DTLS server
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	dtlsConn, err := dtls.DialWithContext(ctx, "udp", addr, config)
	util.Check(err)
	fmt.Println("Connected to hue entertainment API!")
	statusbuff := make(chan []LightPacket, 100)

	return dtlsConn, statusbuff, cancel

}

// streamloop starts the event driven connection loop
func streamloop(interval uint, connection *dtls.Conn, channel chan []LightPacket) {
	for {
		status := <-channel
		switch status[0].LightID {
		case 0:
			return
		default:
			connection.Write(createpacket(status))
			time.Sleep(100 * time.Millisecond)
		}

	}
}

func startsession(hubip string, authcode string, lightgroup string) {
	initstream := Apireq{
		Stream: ApiActive{Active: true},
	}

	client := &http.Client{}

	json1, err := json.Marshal(initstream)
	if err != nil {

		panic(err)
	}

	req, err := http.NewRequest(http.MethodPut, "http://"+hubip+"/api/"+authcode+"/groups/"+lightgroup, bytes.NewBuffer(json1))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)

	if err != nil {
		panic(err)
	}
	fmt.Println("Server responded with status code: ", resp.StatusCode)
}

func createpacket(Lights []LightPacket) []byte {
	protoc := []byte("HueStream")
	version := []byte{0x01, 0x00}  // Protocol Version 1.0
	seqnum := []byte{0x07}         // Squence Number (Currently not used)
	reserve1 := []byte{0x00, 0x00} // Reserved Bytes
	colormode := []byte{0x00}      // RGB Color Mode
	reserve2 := []byte{0x00}

	protoct := append(protoc[:], version[:]...)
	protoct = append(protoct[:], seqnum[:]...)
	protoct = append(protoct[:], reserve1[:]...)
	protoct = append(protoct[:], colormode[:]...)
	protoct = append(protoct[:], reserve2[:]...)

	for _, light := range Lights {
		lightID := []byte{0x00, 0x00, light.LightID}
		print(uint16(light.Red) << 8)
		red := make([]byte, 2)
		binary.BigEndian.PutUint16(red, uint16(light.Red) * 256 + uint16(light.Red))
		green := make([]byte, 2)
		binary.BigEndian.PutUint16(green, uint16(light.Green) * 256 + uint16(light.Green))
		blue := make([]byte, 2)
		binary.BigEndian.PutUint16(blue, uint16(light.Blue) * 256 + uint16(light.Blue))

		protoct = append(protoct[:], lightID[:]...)
		protoct = append(protoct[:], red[:]...)
		protoct = append(protoct[:], green[:]...)
		protoct = append(protoct[:], blue[:]...)
	}
	fmt.Printf("%v \n", protoct)
	return protoct
}
