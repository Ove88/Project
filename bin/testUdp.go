package main

import (
	"com/udp"
	"time"
)

func main() {
	udpSend_ch := make(chan udp.UdpPacket, 5)
	udpReceive_ch := make(chan udp.UdpPacket, 5)
	udp.Init(16000, 15000, udpReceive_ch, udpSend_ch)
	go func(udpReceive_ch chan udp.UdpPacket) {
		for {
			p := <-udpReceive_ch
			println(string(p.Data))
		}
	}(udpReceive_ch)
	for {
		udpSend_ch <- udp.UdpPacket{"broadcast", []byte("Test")}
		time.Sleep(time.Second)
	}

}
