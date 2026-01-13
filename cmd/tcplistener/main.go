package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func main() {
	listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("connection accepted")
		lines := getLinesChannel(conn)
		for line := range lines {
			fmt.Println(line)
		}
		fmt.Println("connection closed")
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)
	go func() {
		defer close(lines)
		defer f.Close()

		buffer := make([]byte, 8)
		var currentLine string

		for {
			n, err := f.Read(buffer)
			data := string(buffer[:n])

			if n > 0 {
				parts := strings.Split(data, "\n")
				for i, part := range parts {
					if i == len(parts)-1 {
						currentLine += part
						break
					}
					currentLine += part
					lines <- currentLine
					currentLine = ""
				}
			}

			if err == io.EOF {
				break
			}

			if err != nil {
				log.Fatal(err)
			}
		}

		if currentLine != "" {
			lines <- currentLine
		}
	}()

	return lines
}
