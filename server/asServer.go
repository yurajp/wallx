package server

import (
	"encoding/json"
	"fmt"
	"net"
	"github.com/yurajp/bridge/config"
  "github.com/yurajp/bridge/ascod"
  "github.com/yurajp/bridge/symcod"
  
)

type KeyResp struct {
	Rand string
	Pub  ascod.PubKey
}

type PassMode struct {
  Password string
  Mode string
}

var port string

func SecureHandle(conn net.Conn) {
  // getting client random
	rndBuf := make([]byte, 512)
	n, err := conn.Read(rndBuf[:])
	if err != nil {
		fmt.Printf("cannot read random: %s", err)
		return
	}
	rand := string(rndBuf[:n])
	// generating keys 
	pub, priv, err := ascod.GenerateKeys()
	if err != nil {
		fmt.Printf("cannot generate keys: %s", err)
		return
	}
	// create KeyResp for client
	kRs := ascod.NewKeyResp(rand, pub, priv)
	// sending KeyResp json
	js, err := json.Marshal(kRs)
	if err != nil {
		fmt.Printf("cannot convert keyResp: %s", err)
		return
	}
	_, er := conn.Write(js)
	if er != nil {
		fmt.Printf("cannot send keyResp: %s", er)
		return
	}
	// getting struct with encrypted password for symmetric encoding and /
	// mode (files|text) encrypted by this password
	passMdBuf := make([]byte, 1024)
	m, err := conn.Read(passMdBuf[:])
	if err != nil {
		fmt.Printf("cannot receive passMode: %s", err)
		return
	}
	var encPM PassMode
	err = json.Unmarshal(passMdBuf[:m], &encPM)
	// handling the struct and getting password and mode
	decPwd := ascod.SrvDecodeString(encPM.Password, priv)
	mode := symcod.SymDecode(encPM.Mode, decPwd)
  sOk := ascod.SrvEncodeString("OK", priv)
  
  // define further action
	if mode == "files" {
	  conn.Write([]byte(sOk))
    go GetFiles(conn)
	} else if mode == "text" {
    conn.Write([]byte(sOk))
	  go GetText(conn, decPwd)
	} else {
	  conn.Write([]byte("error"))
	  conn.Close()
	}
}	
	
func AsServer() { 
  port = config.Conf.Server.Port
	listen, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Printf("Failed to establish connection: %s\n", err)
	}
	fmt.Printf("\n    SERVER started on %s\n ", port)
	for {
	  // select {
	  // case <-stop:  
	  //   break
	  // default:
		conn, err := listen.Accept()
		if err != nil {
			fmt.Printf("connection failed ...continue...\n")
			continue
		}
		go SecureHandle(conn)
	}
}
