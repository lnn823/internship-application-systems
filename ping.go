package main

import (
	"bytes"
	"encoding/binary"
	"flag"
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

var (
	packageTrans int
	packageRecv int
	durations []float64
	ipv6 bool
	ttlLimit int
	count int
	wait int
	timeout int
	help bool
	size int
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
	var IPconn *net.IPConn
	var err error
	if ipv6 {
		IPconn, err = net.DialIP("ip6:ipv6-icmp", nil, ipAddr)
		icmp.Type = 128
	} else {
		IPconn, err = net.DialIP("ip4:icmp", nil, ipAddr)
	}
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return err
	}
	defer IPconn.Close()
	buffer := new(bytes.Buffer)
	_ = binary.Write(buffer, binary.BigEndian, icmp)
	for i :=0 ; i < size; i++ {
		buffer.WriteByte(0xff)
	}
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
	ttl := int(rBuffer[8])
	if ttl <= ttlLimit {
		fmt.Printf("%d bytes from %s: icmp_seq=%d ttl=%d time=%.3f ms\n", len(rBuffer[20:length]), ipAddr.String(), icmp.Seq, ttl, dTime)
	} else {
		fmt.Printf("%d bytes from %s: icmp_seq=%d ttl=%d time=%.3f ms. TIME EXCEEDED\n", len(rBuffer[28:length]), ipAddr.String(), icmp.Seq, ttl, dTime)
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Not enough arguments")
		os.Exit(0)
	}
	flag.BoolVar(&ipv6, "6", false, "IPv6")
	flag.IntVar(&ttlLimit, "m", math.MaxInt32, "TTL limit")
	flag.IntVar(&count, "c", math.MaxInt32, "Stop after sending (and receiving) count ECHO_RESPONSE packets.")
	flag.IntVar(&wait, "i", 1, "Wait wait seconds between sending each packet.  The default is to wait for one second between each packet.")
	flag.IntVar(&timeout, "t", math.MaxInt32, "Specify a timeout, in seconds, before ping exits regardless of how many packets have been received.")
	flag.BoolVar(&help, "h", false, "This help")
	flag.IntVar(&size, "s", 56, "Specify the number of data bytes to be sent.")
	flag.Parse()
	if help {
		fmt.Println("Ping GO version\n" +
			"Usage: sudo ./ping [-6h] [-m ttl] [-c count] [-i wait] [-t timeout] [-s size] dest_ip_addr\n" +
			"\n" +
			"Options: ")
		flag.PrintDefaults()
		os.Exit(0)
	}
	destAddr := os.Args[len(os.Args) - 1]
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
	fmt.Printf("PING %s (%s): %d data bytes\n", destAddr, ipAddr.String(), size)
	for i := 0; i < int(math.Min(float64(count), float64(timeout / wait))); i++ {
		err = sendRequest(ipAddr, i)
		if err != nil {
		}
		for i :=0; i < wait; i++ {
			time.Sleep(time.Second)
		}
	}
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