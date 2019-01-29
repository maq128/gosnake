package main

import (
	"fmt"
	"gosnake-server/comm"
	"log"
	"net"

	"github.com/golang/protobuf/proto"
)

func main() {
	pc, err := net.ListenPacket("udp4", ":6688")
	if err != nil {
		log.Fatal("listenPacket:", err)
	}
	defer pc.Close()

	addr := pc.LocalAddr()
	log.Println("Serve UDP at:", addr)

	doneChan := make(chan error, 1)
	buffer := make([]byte, 1024)

	// Given that waiting for packets to arrive is blocking by nature and we want
	// to be able of canceling such action if desired, we do that in a separate
	// go routine.
	go func() {
		for {
			n, addr, err := pc.ReadFrom(buffer)
			if err != nil {
				doneChan <- err
				return
			}

			fmt.Printf("packet-received: bytes=%d from=%s\n", n, addr.String())

			up := &comm.Up{}
			err = proto.Unmarshal(buffer[:n], up)
			if err != nil {
				log.Println("bad receive data:", err)
				continue
			}

			log.Println("receive:", up)

			// deadline := time.Now().Add(time.Second * 2)
			// err = pc.SetWriteDeadline(deadline)
			// if err != nil {
			// 	doneChan <- err
			// 	return
			// }

			// n, err = pc.WriteTo(buffer[:n], addr)
			// if err != nil {
			// 	doneChan <- err
			// 	return
			// }

			// fmt.Printf("packet-written: bytes=%d to=%s\n", n, addr.String())
		}
	}()

	select {
	case err = <-doneChan:
	}
}
