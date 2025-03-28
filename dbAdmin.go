//uds client to test ApiDbServer

package main

import (
    "fmt"
    "net"
    "os"
	"bufio"
//	"log"
	"strings"
)

const SockAddr = "/tmp/api.db"
var cmdList = [...]string {"chk","add","del","upd","lis"}

func main() {

	var cmd string
	var msg string
	var n2 int
	rd := bufio.NewReader(os.Stdin)
    conn, err := net.Dial("unix", SockAddr)
    if err != nil {
        fmt.Println("Error connecting:", err.Error())
        os.Exit(1)
    }
    defer conn.Close()

	buf := make([] byte, 1024)

	for i:=0; i<10; i++ {
		fmt.Printf("Enter msg>")
		inplin,_,err := rd.ReadLine()
		if err != nil {fmt.Printf("invalid input\n"); continue;}
		inpList := strings.Split(string(inplin)," ")
		n := len(inpList)
		if n == 0 || n>3 { fmt.Printf("incorrect args: %d!\n", n); continue;}
		cmd = inpList[0]
//		n,err := fmt.Scanf("%s %s %s\n",  &cmd, &key, &val)
//		if err !=nil {fmt.Printf("error -- scan message: %v\n", err); continue;}
//		fmt.Printf("  %s %s %s\n", cmd, key, val)

		switch cmd {
		case "chk":
			if n!=3 {
				fmt.Printf("incorrect args!\n")
				goto loop
			}
			msg ="chk " + inpList[1] + ":" + inpList[2]

		case "get":
			if n!=2 {
				fmt.Printf("incorrect args!\n")
				goto loop
			}
			msg = "get " + inpList[1]

		case "add":
			if n!=3 {
				fmt.Printf("incorrect args!\n")
				goto loop
			}
			msg ="add " + inpList[1] + ":" + inpList[2]

		case "lis":
			if n!=1 {
				fmt.Printf("incorrect args!\n")
				goto loop
			}
			msg = "lis"

		case "del":
			if n!=2 {
				fmt.Printf("incorrect args!\n")
				goto loop
			}
			msg = "del " + inpList[1]
//			msg = "del " + key

		case "upd":
			if n!=3 {
				fmt.Printf("incorrect args!\n")
				goto loop
			}
			msg ="upd " + inpList[1] + ":" + inpList[2]
//			msg ="upd " + key + ":" + val

		case "exit":
			// send eof
			fmt.Printf("exiting\n")
			goto fin

		default:
			fmt.Printf("invalid cmd: %s\n", cmd)
			goto loop

		}

		fmt.Printf("msg>%s\n", msg)
    	_, err = conn.Write([]byte(msg))
    	if err != nil {fmt.Printf("error -- sending message: %v\n", err); continue;}

	    // Wait for response
    	n2, err = conn.Read(buf)
    	if err != nil {
			fmt.Println("Error -- receiving response:", err.Error())
		} else {
    		fmt.Printf("Received response[%d]: %s\n",n2, string(buf[0: n2]))
		}
loop:
	}

fin:
}
