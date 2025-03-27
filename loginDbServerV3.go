// loginDbServer
// uds server that manages web sites login.
//
// author: prr, azulsoftware
// copyright: 2025 prr, azulsoftware
//
// v2 add gracefull shutdown
// v3 wait for dbHandler
//

package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"io"
	"os/signal"
	"syscall"
	"sync"
	"time"
	"strings"
	"bytes"

//    pgdb "github.com/prr123/pogreb/gogLib"
	pgdb "github.com/akrylysov/pogreb"

)

const SockAddr = "/tmp/api.db"
const dbFil = "/home/peter/cloud/db/login.db"

type DbFunc func(inp []byte, db *pgdb.DB) (out []byte, err error)
var hmap = make(map[string]DbFunc)

func dbHandler(conn net.Conn, db *pgdb.DB, fini chan struct{}, wg sync.WaitGroup) {

	var buf [1024]byte

	wg.Add(1)
	defer wg.Done()

	for {
		select {
		case <-fini:
			goto loop
		default:
			oneS := time.Now().Add(1 * time.Second)
			conn.SetReadDeadline(oneS)
			n, err := conn.Read(buf[:])
			if err != nil {
				if os.IsTimeout(err) {continue}

				if err ==  io.EOF {
					log.Printf("conn terminated by EOF\n")
					conn.Close()
					return
				}
				if strings.Contains(err.Error(), "connection reset by peer") {
					log.Printf("conn reset by peer\n")
					conn.Close()
					return
				}

				errmsg:= fmt.Sprintf("error -- conn Read: %v", err)
				SendErrMsg(conn, errmsg)
				continue
			}
			fmt.Printf("read [%d]: %s\n",n, buf[:n])

		if n < 3 {
			errmsg := fmt.Sprintf("error -- buf length too short\n")
			SendErrMsg(conn, errmsg)
			continue
		}

		cmd:=string(buf[:3])
		fmt.Printf("info -- cmd: %s\n", cmd)
    	fn, ok := hmap[cmd]
    	if !ok {
			errmsg := fmt.Sprintf("error -- invalid cmd: %s\n",cmd)
			SendErrMsg(conn, errmsg)
			continue
		}

		msg, err := fn(buf[3:n], db)
		if err != nil {
			errmsg := fmt.Sprintf("error -- cannot get msg for cmd %s: %v\n", cmd, err)
			SendErrMsg(conn, errmsg)
		}
		fmt.Printf("dbg -- msg [%d]: %s\n",len(msg), msg)
		_, err = conn.Write(msg)
    	if err != nil {
			log.Printf("error -- write msg: %v\n", err)
			conn.Close()
			continue
		}
		if cmd=="end" {
			conn.Close();
			db.Close();
			return;
		}
		}// select
	} // loop
loop:

}

func SendErrMsg(conn net.Conn, errmsg string) {
	log.Printf("%s\n", errmsg)
	_, err := conn.Write([]byte(errmsg))
	if err != nil {
		log.Printf("error - write: %v\n", errmsg, err)
		conn.Close()
		os.Exit(1)
	}
	return
}


func dbcheck(inp []byte, db *pgdb.DB) (out []byte, err error) {
	fmt.Printf("dbg -- dbcheck: %s\n", inp)
	val, err := db.Get(inp)
	if err != nil {return out, fmt.Errorf("db read: %v", err)}
	log.Printf("read key: %s val: %s", inp, val)
	return val, nil
}

func dbadd(inp []byte, db *pgdb.DB) (out []byte, err error) {
	fmt.Printf("dbg -- dbadd: %s\n", inp)
	idx:= bytes.IndexByte(inp, ':')
	if idx == -1 {return out, fmt.Errorf("inp has no colon!")}
fmt.Printf("add %s %s\n",inp[:idx], inp[idx+1:])
	err = db.Put(inp[:idx], inp[idx+1:])
	if err != nil {return out, fmt.Errorf("db put: %v", err)}
	out = []byte("add ok")
	return out, nil
}

func dblist(inp []byte, db *pgdb.DB) (out []byte, err error) {
	lin:= make([]byte, 0, 128)
	fmt.Printf("dbg -- dblist: %s\n", inp)
	it := db.Items()

	for ic:= 1; ;ic++ {
    	key, val, err := it.Next()
    	if err == pgdb.ErrIterationDone {break}
    	if err != nil { return out, fmt.Errorf("db get [%d]: %v",ic, err)}
		fmt.Printf("list line [%d]: %d:%s %d:%s",ic,len(key), key, len(val), val)
		lin = append(key,':')
		lin = append(lin,val...)
		lin = append(lin,'\n')
		fmt.Printf("dbg -- line[%d]: %s", len(lin), lin)
		out = append(out,lin...)
		fmt.Printf("dbg -- out[%d] %s\n", len(out), out)
	}
//	out = []byte("check ok")
	return out, nil
}

func dbupd(inp []byte, db *pgdb.DB) (out []byte, err error) {
	fmt.Printf("dbg -- dbupd: %s\n", inp)
	out = []byte("check ok")
	return out, nil
}

func dbdel(inp []byte, db *pgdb.DB) (out []byte, err error) {
	fmt.Printf("dbg -- dbdel: %s\n", inp)
	err = db.Delete(inp)
	if err != nil {return out, fmt.Errorf("db del: %v", err)}
	out = []byte("ok")
	return out, nil
}

func dbend(inp []byte, db *pgdb.DB) (out []byte, err error) {
	fmt.Printf("dbg -- dbend: %s\n", inp)
	out = []byte("check ok")
	return out, nil
}

func procSig(db *pgdb.DB, fini chan struct{}, l net.Listener) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// blocks until a  signal is received
    sig := <-c
    fmt.Printf("Received signal: %v\n", sig)
    // Perform cleanup or graceful shutdown
	l.Close()
	fini<- struct{}{}
	db.Close()
    os.Exit(0)
}


func main() {

	var wg sync.WaitGroup
	if err := os.RemoveAll(SockAddr); err != nil {
		log.Fatalf("error -- cannot remove socket file: %v!\n", err)
	}

	hmap["chk"] = dbcheck
	hmap["lis"] = dblist
	hmap["add"] = dbadd
	hmap["upd"] = dbupd
	hmap["del"] = dbdel
	hmap["end"] = dbend

    db, err := pgdb.Open("pogreb.test", nil)
    if err != nil { log.Fatalf("error -- opend db: %v\n", err)}
    defer db.Close()

	fini := make(chan struct{})
	l, err := net.Listen("unix", SockAddr)
	if err != nil {log.Fatalf("error -- listen error: %v\n", err)}
	defer l.Close()

	go procSig(db, fini, l)

	log.Printf("listening on %s\n", SockAddr)

	for {
		conn, err := l.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {break}
			log.Printf("error -- accepting conn: %v\n", err)
			break
		}
		log.Println("accepted a conn!")
		go dbHandler(conn, db, fini, wg)
	}
	fmt.Println("waiting on dbHandlers to shut down")
	wg.Wait()
	fmt.Println("shut down")
}
