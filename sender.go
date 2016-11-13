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
	buf, _ := sendTime.MarshalText()
	
	//display sent time
	fmt.Printf("Packet was sent at: %s\n", sendTime)
	
	//send to socket
    conn.Write(buf)
}