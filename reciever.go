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
	sentTime := new(time.Time)
	sentTime.UnmarshalText(buf[:n])
	
	// find diffrence between sent time and recieved Time
	travelTime := recieveTime.Sub(*sentTime)
	
	// display them
	fmt.Printf("Packet was sent at: %s\n", sentTime)
	fmt.Printf("Traveling time: %s\n", travelTime)


}