package main

import "net"
import "fmt"
import "time"

func main() {

  fmt.Println("server is running")

  // listen all
  ln, _ := net.Listen("tcp", ":8081")

  for {
        conn, err := ln.Accept()
        if err != nil {
            continue
        }
        handleClient(conn) // running until ctrl-c
        conn.Close() // close
    }
    
}

func handleClient(conn net.Conn) {

    var buf [512]byte
	
	// find size of buf
    n, _ := conn.Read(buf[0:])
	
	// get current time
	recieveTime := time.Now()
	
	// create new time and set time to what is in buf
	sentTime := bytesToTime(buf[:n])
	
	// find diffrence between sent time and recieved Time
	travelTime := recieveTime.Sub(sentTime)
	
	// display them
	fmt.Printf("Packet was sent at: %s\n", sentTime)
	fmt.Printf("Traveling time: %s\n", travelTime)


}

func bytesToTime(b []byte) time.Time {
	var nsec int64
	for i := uint8(0); i < 8; i++ {
		nsec += int64(b[i]) << ((7 - i) * 8)
	}
	return time.Unix(nsec/1000000000, nsec%1000000000)
}
