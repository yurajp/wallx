package web

import (
  "net/http"
  "html/template"
  "os/exec"
  "fmt"
  "embed"
  "regexp"
  "strconv"
  "strings"
  "github.com/yurajp/bridge/config"
  "github.com/yurajp/bridge/client"
  "github.com/yurajp/bridge/server"
  "github.com/yurajp/bridge/database"
)

var (
  //go:embed files
  webDir embed.FS
  fs http.Handler
  hmTmpl *template.Template
  srTmpl *template.Template
  clTmpl *template.Template
  blTmpl *template.Template
  lkTmpl *template.Template
  lvTmpl *template.Template
  Cmode = make(chan string, 1)
  Q = make(chan struct{}, 1)
  SrvUp bool
)


func init() {
  fs = http.FileServer(http.FS(webDir))
  hmTmpl, _ = template.ParseFS(webDir, "files/hmTmpl.html")
  srTmpl, _ = template.ParseFS(webDir, "files/srTmpl.html")
  clTmpl, _ = template.ParseFS(webDir, "files/clTmpl.html")
  blTmpl, _ = template.ParseFS(webDir, "files/blank.html")
  lkTmpl, _ = template.ParseFS(webDir, "files/linkQuery.html")
  lvTmpl, _ = template.ParseFS(webDir, "files/linkView.html")
}

func home(w http.ResponseWriter, r *http.Request) {
  if SrvUp {
    port := config.Conf.Server.Port
    serv := fmt.Sprintf("server is runing on %s", port)
    srTmpl.Execute(w, serv)
  } else {
    hmTmpl.Execute(w, nil)
  }
}

func serverLauncher(w http.ResponseWriter, r *http.Request) {
  if !SrvUp {
    Cmode <-"server"
  }
  SrvUp = true
  srTmpl.Execute(w, server.ToWeb)
}

func textLauncher(w http.ResponseWriter, r *http.Request) {
  Cmode <-"text"
  for {
    select {
      case <-client.Res:
      clTmpl.Execute(w, client.Result)
      return
      default:
    }
  }
}

func filesLauncher(w http.ResponseWriter, r *http.Request) {
  Cmode <-"files"
  for {
    select {
    case <-client.Res:
      clTmpl.Execute(w, client.Result)
      return
      default:
    }
  }
}

func linkView(w http.ResponseWriter, r *http.Request) {
  if r.Method == http.MethodGet {
    lkTmpl.Execute(w, nil)
  }
  if r.Method == http.MethodPost {
    err := r.ParseForm()
    if err != nil {
      fmt.Println(err)
      http.Error(w, err.Error(), http.StatusBadRequest)
    }
    lview, err := database.MakeView()
    if err != nil {
      fmt.Println(err)
      http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    val := r.FormValue("query")
    renum := regexp.MustCompile(`^\d{1,}$`)
    if renum.MatchString(val) {
      num, _ := strconv.Atoi(val)
      lvTmpl.Execute(w, lview[:num])
    } else {
      fnd := []database.LinkView{}
      for _, lv := range lview {
        if strings.Contains(strings.ToLower(lv.Title), strings.ToLower(val)) {
          fnd = append(fnd, lv)
        }
      }
      if len(fnd) == 0 {
        fnd = []database.LinkView{database.LinkView{"", "NOTHING FIND", ""}}
      }
      lvTmpl.Execute(w, fnd)
    }
  }
}

func quit(w http.ResponseWriter, r *http.Request) {
  err := blTmpl.Execute(w, "Bridge closed")
  if err != nil {
    fmt.Println(err)
  }
  Q <-struct{}{}
}

func Launcher() {
  mux := http.NewServeMux()
  mux.HandleFunc("/", home)
  mux.HandleFunc("/server", serverLauncher)
  mux.HandleFunc("/text", textLauncher)
  mux.HandleFunc("/files", filesLauncher)
  mux.HandleFunc("/links", linkView)
  mux.HandleFunc("/quit", quit)
  mux.Handle("/files/", fs)
  hsrv := &http.Server{Addr: ":" + config.Conf.WebPort, Handler: mux}
  
  go hsrv.ListenAndServe()
  err := exec.Command("xdg-open", "http://localhost:" + config.Conf.WebPort).Run()
  if err != nil {
    fmt.Println("xdg-open: ", err)
  }
}