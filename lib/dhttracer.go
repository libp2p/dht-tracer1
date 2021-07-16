package dhttracer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	cid "github.com/ipfs/go-cid"
	dhtnode "github.com/libp2p/dht-tracer1/lib/dhtnode"
	peer "github.com/libp2p/go-libp2p-core/peer"
)

var Version = "1.0.0"

type cancelCtx struct {
	context.Context

	cancel context.CancelFunc
}

type (
	Command = string
	Key     = string
)

var (
	CmdExit         = "exit"
	CmdReset        = "reset"
	CmdPutValue     = "put-value"
	CmdGetValue     = "get-value"
	CmdAddProvider  = "add-provider"
	CmdGetProviders = "get-providers"
	CmdFindPeer     = "find-peer"
	CmdPing         = "ping"
)

var CtrlCmds = []string{
	CmdExit,
	CmdReset,
}

var QueryCmds = []string{
	CmdPutValue,
	CmdGetValue,
	CmdAddProvider,
	CmdGetProviders,
	CmdFindPeer,
	CmdPing,
}

var AllCmds = append(QueryCmds, CtrlCmds...)

// func init() {
//   AllCmds =
// }

func cmdInGroup(c string, g []string) bool {
	for _, s := range g {
		if c == s {
			return true
		}
	}
	return false
}

type Tracer struct {
	NodeCfg dhtnode.NodeCfg

	Node *dhtnode.Node
	ctx  cancelCtx

	sync.RWMutex
}

func NewTracer(cfg dhtnode.NodeCfg) *Tracer {
	return &Tracer{
		NodeCfg: cfg,
	}
}

func (t *Tracer) Start() error {
	t.Lock()
	defer t.Unlock()

	// setup the dht node's context
	ctx, cancel := context.WithCancel(context.Background())
	t.ctx.Context = ctx
	t.ctx.cancel = cancel

	// create node
	var err error
	t.Node, err = dhtnode.NewNode(t.NodeCfg)
	if err != nil {
		return err
	}

	err = dhtnode.Bootstrap(t.Node, t.NodeCfg.Bootstrap)
	if err != nil {
		return err
	}
	return nil
}

func (t *Tracer) Stop() {
	panic("not yet implemented")
	// t.Lock()
	// defer t.Unlock()

	// if t.ctx.cancel != nil {
	//   t.ctx.cancel()
	//   t.ctx.cancel = nil
	// }

	// stop node
	// delete node
}

func (t *Tracer) RunQuery(cmd Command, key Key, vals ...string) (io.Reader, error) {
	t.RLock()
	defer t.RUnlock()

	ctx, cancel := context.WithCancel(t.ctx)
	defer cancel()

	if len(key) < 1 {
		return nil, fmt.Errorf("please enter a Key")
	}

	// run query on node, return the result or closer peers.
	switch cmd {
	case CmdPutValue:
		if len(vals) < 1 {
			return nil, fmt.Errorf("PutValue takes in 1 argument")
		}
		err := t.Node.DHT.PutValue(ctx, key, []byte(vals[0]))
		if err != nil {
			return nil, err
		}
		s := fmt.Sprintf("put %v %v", key, vals[0])
		return strings.NewReader(s), nil
	case CmdGetValue:
		val, err := t.Node.DHT.GetValue(ctx, key)
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(val), nil
	case CmdAddProvider:
		c, err := cid.Decode(key)
		if err != nil {
			return nil, err
		}
		err = t.Node.DHT.Provide(ctx, c, true)
		if err != nil {
			return nil, err
		}
		s := fmt.Sprintf("added self as provider for %v", key)
		return strings.NewReader(s), nil
	case CmdGetProviders:
		c, err := cid.Decode(key)
		if err != nil {
			return nil, err
		}
		pvs := t.Node.DHT.FindProvidersAsync(ctx, c, 10)
		pr, pw := io.Pipe()
		go func() {
			for {
				pv, ok := <-pvs // should respect context
				if !ok {        // channel closed
					return
				}
				pw.Write([]byte(pv.ID.String())) // TODO handle write err?
			}
		}()
		return pr, nil
	case CmdFindPeer:
		pid, err := peer.Decode(key)
		if err != nil {
			return nil, err
		}
		ai, err := t.Node.DHT.FindPeer(ctx, pid)
		if err != nil {
			return nil, err
		}
		return strings.NewReader(ai.String()), nil
	case CmdPing:
		pid, err := peer.Decode(key)
		if err != nil {
			return nil, err
		}
		t1 := time.Now()
		err = t.Node.DHT.Ping(ctx, pid)
		if err != nil {
			return nil, err
		}
		d := time.Since(t1)
		s := fmt.Sprintf("ping time: %v", d)
		return strings.NewReader(s), nil
	default:
		return nil, fmt.Errorf("unknown command")
	}
}

func (t *Tracer) Reset() (io.Reader, error) {
	// t.Stop()
	err := t.Start()
	if err != nil {
		return nil, err
	}
	return strings.NewReader("restarted"), nil
}

// func (t *Tracer) Repl(rw io.ReadWriter) {
//   out := func(s interface{}) {
//     switch st := s.(type) {
//     case []byte:
//       rw.Write(st)
//     case string:
//       rw.Write([]byte(st))
//     case io.Reader:
//       io.Copy(rw, st)
//     default:
//       panic("cannot print type", st)
//     }
//   }

//   checkErr := func(err error) {
//     if err == nil {
//       return false
//     }
//     rw.Write(fmt.Sprintf("error: %v", err))
//     return true
//   }

//   readCmd := func(r io.Reader) (cmd int32, args []string, err error) {
//     args := line.Split(" ")
//     if len(args) < 2 {
//       return nil, errors.New("usage: <command> <arg>...")
//     }

//     cmdType, found := TypeStrToInt[args[0]]
//     if !found {
//       return nil, errors.New("unrecognized command: ", args[0])
//     }

//     return cmdType, args[1:], nil
//   }

//   r := bufio.NewReader(rw)
//   for { // repl loop
//     line, err := r.ReadString('\n')
//     if checkErr(err) {
//       return err
//     }

//     cmd, args, err := readCmd(line)
//     if checkErr(err) {
//       continue
//     }

//     switch {
//     case cmd == CmdExit:
//       rw.Write([]byte("exiting\n"))
//     case cmd == CmdReset:
//       err = t.Reset()
//     case cmdInGroup(cmd, QueryCmds):
//     }

//   }
// }
