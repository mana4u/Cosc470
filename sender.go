package main

import "net"
import "fmt"
import "time"

func main() {

    //connect socket
    conn, _ := net.Dial("tcp", "127.0.0.1:8081")	
	
	//get current time
	sendTime := time.Now()
	
	//encoding time to byte
	buf := timeToBytes(sendTime)
	
	//display sent time
	fmt.Printf("Packet was sent at: %s\n", sendTime)
	
	//send to socket
    conn.Write(buf)
}

func timeToBytes(t time.Time) []byte {
	nsec := t.UnixNano()
	b := make([]byte, 8)
	for i := uint8(0); i < 8; i++ {
		b[i] = byte((nsec >> ((7 - i) * 8)) & 0xff)
	}
	return b
}
