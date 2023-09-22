package server

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"time"
	"github.com/yurajp/bridge/config"
	"github.com/yurajp/bridge/database"
	"github.com/yurajp/bridge/ascod"
	"github.com/yurajp/bridge/symcod"
)

var (
  ToWeb []string
  tdir string
  fdir string
)

type Letter struct {
	Text string
	SHA  string
}

type Upfile struct {
	Fname string
	Fsize int
	Isdir bool
}


func anyBytes(b int) string {
	switch {
	case b < 1000:
		return fmt.Sprintf("%d B", b)
	case b < 1000000:
		k := float64(b) / 1000
		return fmt.Sprintf("%.2f kB", k)
	default:
		m := float64(b) / 1000000
		return fmt.Sprintf("%.2f mB", m)
	}
}

func toWeb(s string) {
  ToWeb = append(ToWeb, s)
}

func writeMonthFile(dir, text string) error {
  fname := dir + "/" + time.Now().Format("2006-01")
  f, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0640)
  if err != nil {
    return fmt.Errorf("Cannot open flle for writing: %w", err)
  }
  defer f.Close()
  f.WriteString(text + "\n")
  return nil
}


func GetText(conn net.Conn, pw string) {
  tdir = config.Conf.Server.TxtDir
  defer conn.Close()
  msg := "OK"
  printer := func(m string, err error) {
    fmt.Printf("%s: %s", msg, err)
  }
  send := func(m string) {
    conn.Write([]byte(msg))
  }

  rd := bufio.NewReader(conn)
	var buf [1024 * 96]byte
	n, err := rd.Read(buf[:])
	if err != nil && err != io.EOF {
		 msg = "Cannot read your text"
		 printer(msg, err)
		 send(msg)
		 return
	}
	var lt Letter
	err = json.Unmarshal(buf[:n], &lt)
	if err != nil {
	  msg = "Cannot unmarshal the letter"
	  printer(msg, err)
	  send(msg)
	  return
	}
	
	text := symcod.SymDecode(lt.Text, pw)
	if text == "" {
	  msg = "Empty text"
	  printer(msg, nil)
	  send(msg)
	  return
	}
	sha := ascod.HashStr(text)
	if sha != lt.SHA {
	  msg = "Hashsums are not matched"
	  printer(msg, err)
	  send(msg)
	  return
	}
	
	go func() {
	  err := writeMonthFile(tdir, text)
	  if err != nil {
	    m := "Cannot write text in file: "
	    fmt.Println(m, err)
	  }
	  tw := "@ A letter was received and stored"
    fmt.Println(tw)
    toWeb(tw)
	}()
	
	x, err := database.LinkScanner(text)
	if err != nil {
	  msg = "Database error"
	  printer(msg, err)
	  send(msg)
	}
	if x > 0 {
	  sfx := "s are"
	  if x == 1 {
	    sfx = " is"
	  }
	  tw2 := fmt.Sprintf(" @ %d link%s stored", sfx, x)
	  fmt.Println(tw2)
	  toWeb(tw2)
	} else {
	  fmt.Println(" No links")
	}
	send("OK")
}


func GetFiles(conn net.Conn) {
  fdir = config.Conf.Server.FileDir
  defer conn.Close()
	var count = 0
	
	send := func(m string) {
	  conn.Write([]byte(m))
	}
	
	printer := func(m string, err error) {
	  fmt.Printf("%s: %s\n", m, err)
	}
	msg := ""
	for {
		jsf := make([]byte, 1024)
		n, err := conn.Read(jsf[:])
		if n == 0 {
			return
		}
		if err != nil && err != io.EOF {
		  msg = "Cannot read Json"
			printer(msg, err)
			send(msg)
			return
		}
		var u Upfile
		err = json.Unmarshal(jsf[:n], &u)
		if err != nil {
			msg = "Cannot unmarshal Json"
			printer(msg, err)
			send(msg)
			return
		}
		send("ok")
		
		fmt.Printf("\n\t Downloading %s (%s)\n", u.Fname, anyBytes(u.Fsize))
		dname := fdir + "/" + u.Fname
		data := make([]byte, 0)
		
		size, part := 0, 0
		bar := u.Fsize / 24
		for size < u.Fsize {
			tempBuf := make([]byte, 1024 * 4)
			m, err := conn.Read(tempBuf[:])
			if err != nil && err != io.EOF {
				msg = "Cannot read data"
				printer(msg, err)
				send(msg)
				return
			}
			size += m
			part += m
			if part >= bar {
			  fmt.Print(">")
			  part = 0
			}
			data = append(data, tempBuf[:m]...)
			
			if errors.Is(err, io.EOF) {
				break
			}
		}
	  fmt.Print(" ✓")
		err = os.WriteFile(dname, data, 0664)
		if err != nil {
			msg = "Cannot write file"
			printer(msg, err)
			send(msg)
			return
		}
		time.Sleep(time.Millisecond * 150)
		din, err := os.Stat(dname)
		if err != nil {
			msg = "Cannot get Stat"
			printer(msg, err)
			send(msg)
			return
		}
		if u.Fsize != int(din.Size()) {
			msg = "File size error"
			printer(msg, err)
			send(msg)
			return
		}
		if u.Isdir {
		  	wdr, _ := os.Getwd()
		  	os.Chdir(fdir)
		//	ddr, _ := os.Getwd()
		  	unz := exec.Command("unzip", u.Fname)
		  	err = unz.Run()
		  	if err != nil {
		  		fmt.Println("\n\t Cannot unzip file")
		  	}
		  	rmz := exec.Command("rm", u.Fname)
		  	err = rmz.Run()
		  	if err != nil {
		  		fmt.Println("\n\t Cannot delete zipfile")
		  	}
		  	os.Chdir(wdr)
		}
		time.Sleep(time.Millisecond * 150)

		send("ok")

		req := make([]byte, 128)
		i, err := conn.Read(req[:]) 
		if err != nil {
			fmt.Println(err)
			return
		}
		count++
		if string(req[:i]) == "finish" {
			break
		}
		send("ok")
	}
	send("ok")
	ss := "s"
	if count == 1 {
		ss = ""
	}
	fmt.Printf("\n\n\t%d file%s downloaded\n ", count, ss)
}

