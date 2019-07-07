package main

// type Repl struct {
//   rw   io.ReadWriter
//   net  dhtnode.Net
//   cmds map[string]func()
// }

// func NewRepl(net *dhtnode.Net) *Repl {
//   repl := &Repl{}
//   repl.net = net
//   repl.cmds := map[string]func(){
//     "stats": repl.Stats
//   }

// }

// func (repl *Repl) Dispatch(line string) error {

//   cmd := strings.Split(line, " ")
//   if len(cmd) < 1 {
//     return nil
//   }

//   switch cmd[0] {
//   case "stats":
//   case "bootstrap":
//   case "exit":
//   }
// }
