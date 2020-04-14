package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)
type ICMPHead struct {
	Type uint8
	Code uint8
	Checksum uint16
	ID uint16
	Seq uint16
}

type ICMPPackage struct {
	Head ICMPHead
	Payload [64]byte
}

var (
	packageTrans int
	packageRecv int
	durations []float64
)

func setChecksum(icmp ICMPHead) uint16 {
	var buffer bytes.Buffer
	_ = binary.Write(&buffer, binary.BigEndian, icmp)
	bufferByte := buffer.Bytes()
	length := len(bufferByte)
	var sum uint32
	i := 0
	for ; length > 1; length -= 2 {
		sum += uint32(bufferByte[i]) << 8 + uint32(bufferByte[i + 1])
		i += 2
	}
	if length == 1 {
		sum += uint32(bufferByte[i])
	}
	sum += sum >> 16
	return uint16(^sum)
}

func sendRequest(ipAddr *net.IPAddr, seq int) error {
	icmp := ICMPHead{
		Type: 8,
		Code: 0,
		Checksum: 0,
		ID: 1,
		Seq: uint16(seq),
	}
	icmp.Checksum = setChecksum(icmp)
	IPconn, err := net.DialIP("ip4:icmp", nil, ipAddr)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return err
	}
	defer IPconn.Close()
	var buffer bytes.Buffer
	content := [64]byte{0}
	icmpPackage := ICMPPackage{
		Head: icmp,
		Payload: content,
	}
	_ = binary.Write(&buffer, binary.BigEndian, icmpPackage)
	_, err = IPconn.Write(buffer.Bytes())
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return err
	}
	start := time.Now()
	packageTrans++
	_ = IPconn.SetReadDeadline(time.Now().Add(time.Second))
	rBuffer := make([]byte, 65535)
	length, err := IPconn.Read(rBuffer)
	if err != nil {
		fmt.Printf("Request timeout for icmp_seq %d\n", icmp.Seq)
		return err
	}
	end := time.Now()
	packageRecv++
	dTime := float64(end.Sub(start).Nanoseconds()) / 1e6
	durations = append(durations, dTime)
	fmt.Printf("%d bytes from %s: icmp_seq=%d ttl=%d time=%.3f ms\n", len(rBuffer[28:length]), ipAddr.String(), icmp.Seq, int(rBuffer[8]), dTime)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Not enough arguments")
		os.Exit(0)
	}
	destAddr := os.Args[1]
	ipAddr, err := net.ResolveIPAddr("ip", destAddr)
	if err != nil {
		fmt.Printf("ping: cannot resolve %s: Unknown host\n", destAddr)
		return
	}
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT)
	go func() {
		s := <-c
		if s == syscall.SIGINT {
			fmt.Println()
			fmt.Printf("--- %s ping statistics ---\n", destAddr)
			percent := (1 - float64(packageRecv) / float64(packageTrans)) * 100
			fmt.Printf("%d packets transmitted, %d packets received, %.1f%% packet loss\n", packageTrans, packageRecv, percent)
			if packageRecv > 0 {
				min, avg, max, stddev := statistics(durations)
				fmt.Printf("round-trip min/avg/max/stddev = %.3f/%.3f/%.3f/%.3f ms\n", min, avg, max, stddev)
			}
			os.Exit(0)
		}
	}()
	fmt.Printf("PING %s (%s):\n", destAddr, ipAddr.String())
	for i := 0; true; i++ {
		err = sendRequest(ipAddr, i)
		if err != nil {
		}
		time.Sleep(time.Second)
	}
}

func statistics(arr []float64) (float64, float64, float64, float64) {
	max := 0.0
	sum := 0.0
	min := math.MaxFloat64
	for _, value := range arr {
		if value > max {
			max = value
		}
		if value < min {
			min = value
		}
		sum += value
	}
	avg := sum / float64(len(arr))
	sqd := 0.0
	for _, value := range arr {
		sqd += math.Pow(value - avg, 2)
	}
	stddev := math.Sqrt(sqd / float64(len(arr)))
	return min, avg, max, stddev
}