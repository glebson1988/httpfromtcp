package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	fileName := "messages.txt"
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	buffer := make([]byte, 8)
	var currentLine string

	for {
		n, err := file.Read(buffer)
		data := string(buffer[:n])

		if n > 0 {
			parts := strings.Split(data, "\n")
			for i, part := range parts {
				if i == len(parts)-1 {
					currentLine += part
					break
				}
				currentLine += part
				fmt.Printf("read: %s\n", currentLine)
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
		fmt.Printf("read: %s\n", currentLine)
	}
}
