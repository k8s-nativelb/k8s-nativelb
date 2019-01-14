/* UDPDaytimeClient
 */
package main

import (
"net"
"os"
"fmt"
	"time"
)

func main() {
	service := os.Args[1]

	udpAddr, err := net.ResolveUDPAddr("udp4", service)
	checkError(err)

	conn, err := net.DialUDP("udp", nil, udpAddr)
	checkError(err)

	tick := time.Tick(5 * time.Second)
	go timeOut(tick)

	_, err = conn.Write([]byte("anything"))
	checkError(err)

	var buf [512]byte
	n, err := conn.Read(buf[0:])
	checkError(err)

	fmt.Println(string(buf[0:n]))

	os.Exit(0)
}

func timeOut(tick <-chan time.Time) {
	<-tick
	fmt.Println("failed to receive data from server")
	os.Exit(1)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error ", err.Error())
		os.Exit(1)
	}
}
