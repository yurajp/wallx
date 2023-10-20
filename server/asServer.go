package server

import (
	"encoding/json"
	"os"
	"net"
	"github.com/yurajp/bridge/config"
  "github.com/yurajp/bridge/ascod"
  "github.com/yurajp/bridge/symcod"
  "github.com/sirupsen/logrus"
)

type KeyResp struct {
	Rand string
	Pub  ascod.PubKey
}

type PassMode struct {
  Password string
  Mode string
}

var (
	log *logrus.Logger
  port string
  Shutdown = make(chan struct{}, 1)
)

func iserr(s string, err error) bool {
	if err != nil {
		log.WithError(err).Error(s)
		return true
	}
	return false
}

func SecureHandle(conn net.Conn) {
  // getting client random
	rndBuf := make([]byte, 512)
	n, err := conn.Read(rndBuf[:])
	if iserr("Cannot read random", err) {
		return
	}
	rand := string(rndBuf[:n])
	// generating keys 
	pub, priv, err := ascod.GenerateKeys()
	if iserr("Cannot generate keys", err) {
		return
	}
	// create KeyResp for client
	kRs := ascod.NewKeyResp(rand, pub, priv)
	// sending KeyResp json
	js, err := json.Marshal(kRs)
	if iserr("Cannot convert keyResp", err) {
		return
	}
	_, er := conn.Write(js)
	if iserr("Cannot send keyResp", er) {
		return
	}
	// getting struct with encrypted password for symmetric encoding and /
	// mode (files|text) encrypted by this password
	passMdBuf := make([]byte, 1024)
	m, err := conn.Read(passMdBuf[:])
	if iserr("Cannot receive passMode", err) {
		return
	}
	var encPM PassMode
	err = json.Unmarshal(passMdBuf[:m], &encPM)
	if iserr("Cannot unmarshal pass", err) {
		return
	}
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
	log = logrus.New()
	log.Formatter = new(logrus.TextFormatter)
	log.Level = logrus.InfoLevel
	log.Out = os.Stdout
  port = config.Conf.Server.Port
	listen, err := net.Listen("tcp", port)
	if iserr("Failed to establish connection", err) {
		return
	}
	log.Infof("TCP server started on %s", port)
  TCP:
	for {
	  select {
	  case <-Shutdown:  
	    break TCP
    default:
		  conn, err := listen.Accept()
		  if iserr("Connection failed ...continue...\n", err) {
			  continue
		  }
		  go SecureHandle(conn)
	  }
	}
}
