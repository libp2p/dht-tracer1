package dhttracer

import (
  "bufio"
  "context"
  "errors"
  "fmt"
  "html"
  "io"
  "net/http"
  "os"
  "strings"

  lwriter "github.com/ipfs/go-log/writer"
  dhtnode "github.com/libp2p/dht-tracer1/lib/dhtnode"
)

type HTTPServer struct {
  Tracer *Tracer
  Mux    *http.ServeMux
  Server http.Server
}

func NewHTTPServer(t *Tracer, addr string) *HTTPServer {
  s := &HTTPServer{Tracer: t}

  s.Mux = http.NewServeMux()
  s.Mux.HandleFunc("/cmd", s.handleCmd)
  s.Mux.HandleFunc("/events", s.handleEvents)
  s.Mux.HandleFunc("/version", s.handleVersion)
  s.Mux.HandleFunc("/info/switch", s.handleInfoSwitch)

  s.Server.Addr = addr
  s.Server.Handler = s.Mux
  return s
}

func enableCors(w *http.ResponseWriter) {
  (*w).Header().Set("Access-Control-Allow-Origin", "*")
  (*w).Header().Set("Content-Type", "text/event-stream")
}

func (s *HTTPServer) ListenAndServe() error {
  return s.Server.ListenAndServe()
}

func (s *HTTPServer) handleVersion(res http.ResponseWriter, req *http.Request) {
  fmt.Fprintln(os.Stderr, "/version")
  fmt.Fprintf(res, "tracedht version %v\n", Version)
}

func (s *HTTPServer) handleInfoSwitch(res http.ResponseWriter, req *http.Request) {
  fmt.Fprintln(os.Stderr, "/switch")

  // ping all peers twice
  dhtnode.PingPeers(s.Tracer.Node, 2)

  // then print the latency table
  dhtnode.PrintLatencyTable(res, s.Tracer.Node.Host)
}

func (s *HTTPServer) handleCmd(res http.ResponseWriter, req *http.Request) {

  // parse form
  if err := req.ParseForm(); err != nil {
    http.Error(res, fmt.Sprint(err), http.StatusBadRequest)
    return
  }

  // parse command
  line := req.Form.Get("q")
  cmd, args, err := parseCmd(line)
  if err != nil {
    http.Error(res, fmt.Sprint(err), http.StatusBadRequest)
    return
  }

  fmt.Fprintf(os.Stderr, "/cmd %v %v\n", cmd, args)
  // dispatch command
  var r io.Reader
  switch {
  case cmd == CmdExit:
    go s.Server.Close()
    r = strings.NewReader("exiting...")
  case cmd == CmdReset:
    r, err = s.Tracer.Reset()
  case cmdInGroup(cmd, QueryCmds):
    r, err = s.Tracer.RunQuery(cmd, args[0], args[1:]...)
  }
  if err != nil {
    errs := fmt.Sprintf("error: %v", err)
    http.Error(res, errs, http.StatusInternalServerError)
    return
  }

  // print cmd response
  io.Copy(res, r)
}

func (s *HTTPServer) handleEvents(res http.ResponseWriter, req *http.Request) {
  enableCors(&res)
  fmt.Fprintln(os.Stderr, "/events")

  r := eventlogReader(req.Context())
  scanner := bufio.NewScanner(r)
  for scanner.Scan() {
    fmt.Fprintf(res, "data: %s\n\n", scanner.Text())
  }
}

func parseCmd(line string) (cmd string, args []string, err error) {
  line = html.UnescapeString(line)
  line = strings.ReplaceAll(line, "+", " ")

  if len(line) < 1 {
    return "", nil, errors.New("no command provided. use ?q=<cmd>")
  }

  args = strings.Split(line, " ")
  if (args[0] != "reset") && args[0] != "exit" && len(args) < 2 {
    return "", nil, errors.New("command format: <command> <arg>...")
  }

  c := args[0]
  if !cmdInGroup(c, AllCmds) {
    return "", nil, fmt.Errorf("unrecognized command: %v", args[0])
  }

  return c, args[1:], nil
}

func eventlogReader(ctx context.Context) io.Reader {
  r, w := io.Pipe()
  go func() {
    defer w.Close()
    <-ctx.Done()
  }()
  lwriter.WriterGroup.AddWriter(w)
  return r
}
