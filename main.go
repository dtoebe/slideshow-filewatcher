// Author: Daniel Toebe <dtoebe@gmail.com>
//License: All code I have written is licensed under MIT
//	All other code I have not written such as dependacies are licensed under thier
//	respecive License

//slideshow-filewatcher watches a defined directory and sends defined change notification
//	to the slideshow directly via tcp socket
package main

import (
	"flag"
	"log"
	"net"
	"os"
	"strconv"
	"syscall"

	"golang.org/x/exp/inotify"
)

//parseFlags simply parses the flags and returns them
func parseFlags() (string, string, int) {
	path := flag.String("path", "", "Absolute path to file to watch (from root)")
	port := flag.Int("port", 0, "Port to send TCP socket message of change in dir")
	host := flag.String("host", "127.0.0.1", "Host to sent TCP Socket message")

	flag.Parse()

	return *path, *host, *port
}

//checkFlags checks for the needed inputs
func checkFlags(path, host string, port int) bool {
	allGood := true
	if len(path) < 1 {
		log.Println("[ERR] Missing a path: please set a path with the '-path' flag.")
		allGood = false
	} else {
		if !checkPath(path) {
			log.Println("[ERR] Path does not exist: please use an absolute path starting with root '/'")
			allGood = false
		}
	}
	if port <= 0 {
		log.Println("[ERR] Please add a port number to communicate with")
		allGood = false
	}
	return allGood
}

//checkPath takes a path (string) and makes sure it exists
func checkPath(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

//watchPath takes a path (string) and starts the inotify watcher and runs an infinate loop
//	and sends the change message if a defined syscall is run on the watched path
func watchPath(path, host, port string) {
	watcher, err := inotify.NewWatcher()
	if err != nil {
		log.Fatalf("[ERR] Failed to start watcher: %v", err)
	}

	err = watcher.Watch(path)
	if err != nil {
		log.Fatalf("[ERR] Falied to watch path: %v", err)
	}

	for {
		select {
		case ev := <-watcher.Event:
			switch {
			case ev.Mask == syscall.IN_MOVED_FROM:
				socketClient("true", host, port)
			case ev.Mask == syscall.IN_CLOSE_WRITE:
				socketClient("true", host, port)
			case ev.Mask == syscall.IN_DELETE:
				socketClient("true", host, port)
			}
		case err := <-watcher.Error:
			log.Println("Error:", err)
		}
	}
}

//socketClient takes a msg, the hostname and a port (string) sends the msg to the slideshow
func socketClient(msg, host, port string) {
	conn, err := net.Dial("tcp", host+":"+port)
	if err != nil {
		log.Printf("Cannot dial out to server: %v\n", err)
		return
	}
	defer conn.Close()

	_, err = conn.Write([]byte(msg))
	if err != nil {
		log.Printf("Could not send message: %v\n", err)
		return
	}

	reply := make([]byte, 1024)
	_, err = conn.Read(reply)
	if err != nil {
		log.Printf("[ERR] Did not recieve a return message from server: %v\n", err)
	} else {
		log.Printf("[INF] Server replied with: %s", string(reply))
	}

	return
}

func main() {
	path, host, port := parseFlags()
	if !checkFlags(path, host, port) {
		os.Exit(1)
	}

	watchPath(path, host, strconv.Itoa(port))
}
