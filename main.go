package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	fileName := "messages.txt"
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	buffer := make([]byte, 8)

	for {
		n, err := file.Read(buffer)

		if n > 0 {
			fmt.Printf("read: %s\n", string(buffer[:n]))
		}

		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatal(err)
		}
	}
}
