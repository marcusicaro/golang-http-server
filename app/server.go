package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

// Ensures gofmt doesn't remove the "net" and "os" imports above (feel free to remove this!)
var _ = net.Listen
var _ = os.Exit

func handleRequest(conn net.Conn) {

	reader := bufio.NewReader(conn)
	request := ""

	for {

		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Failed to get request string.")
			os.Exit(1)
		}
		line = strings.TrimSpace(line)
		if line == "" {
			break
		}
		request += line + "\n"
	}

	requestHeaders := strings.Split(request, "\n")

	// fmt.Println(requestHeaders)

	requestString := requestHeaders[0]

	requestParts := strings.Split(requestString, " ")

	method := requestParts[0]

	urlParts := strings.Split(requestParts[1], "/")

	contentLength := 0
	for _, line := range requestHeaders {
		if strings.HasPrefix(line, "Content-Length:") {
			parts := strings.Split(line, " ")
			contentLength, _ = strconv.Atoi(parts[1])
			break
		}
	}

	body := make([]byte, contentLength)
	_, err := io.ReadFull(reader, body)
	if err != nil {
		fmt.Println("Failed to read request body.")
		os.Exit(1)
	}

	fmt.Print(string(body))

	if urlParts[1] == "echo" {
		echoWord := strings.Replace(urlParts[2], "/", "", -1)
		contentLength := strconv.Itoa(len(echoWord))
		conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + contentLength + "\r\n\r\n" + echoWord))
	} else if urlParts[1] == "user-agent" {
		userAgent := strings.Split(requestString, ": ")[1]
		contentLength := strconv.Itoa(len(userAgent))
		conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: " + contentLength + "\r\n\r\n" + userAgent))
	} else if urlParts[1] == "files" && method == "GET" {
		dir := os.Args[2]
		fileName := urlParts[2]
		file, err := os.Open(dir + fileName)
		if err != nil {
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		}
		if err == nil {
			defer file.Close()
			fileContent, err := io.ReadAll(file)
			if err != nil {
				conn.Write([]byte("HTTP/1.1 500 internal server error\r\n\r\n"))
			}

			contentLength := strconv.Itoa(len(fileContent))
			conn.Write([]byte(fmt.Sprintf("HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %s\r\n\r\n%s", contentLength, string(fileContent))))
		}
	} else if urlParts[1] == "files" && method == "POST" {
		dir := os.Args[2]
		fileName := urlParts[2]
		err = os.WriteFile(dir+"/"+fileName, body, 0644)
		if err != nil {
			conn.Write([]byte("HTTP/1.1 500 Internal Server Error\r\n\r\n"))
			return
		}
		conn.Write([]byte("HTTP/1.1 201 Created\r\nContent-Length: 0\r\n\r\n"))

	} else if urlParts[1] == "/" {
		conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
	} else {
		conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
	}
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	ln, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleRequest(conn)
	}

}
