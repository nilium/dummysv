// Copyright (C) 2019 Noel Cower.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Program dummysv is a simple echo server for debugging HTTP requests. It responds to all requests
// with a body and status code, along with any headers given on the command line.
package main

import (
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
)

type syncWriter struct {
	w io.Writer
	m sync.Mutex
}

func (w *syncWriter) Write(b []byte) (int, error) {
	w.m.Lock()
	defer w.m.Unlock()
	return w.w.Write(b)
}

var Stdout = syncWriter{w: os.Stdout}

var (
	logger  *log.Logger
	verbose = false
	body    = "OK"
	code    = 200
	headers = http.Header{}
)

func ok(w http.ResponseWriter, req *http.Request) {
	for k, vals := range headers {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(code)
	if len(body) > 0 {
		io.WriteString(w, body)
	}

	if verbose {
		buf, err := httputil.DumpRequest(req, true)
		if err != nil {
			logger.Println("Error dumping request:", err)
			return
		}
		logger.Printf("%s", buf)
	}
}

func main() {
	var addr = "127.0.0.1:8080"
	var netw = "tcp"

	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.StringVar(&body, "r", body, "The `body` to reply with.")
	flag.IntVar(&code, "s", code, "The `status` code to respond with.")
	flag.StringVar(&netw, "n", netw, "The `network` to listen on.")
	flag.StringVar(&addr, "L", addr, "The `address` to listen on.")
	flag.BoolVar(&verbose, "v", verbose, "Whether to log all received requests.")
	flag.Parse()

	logger = log.New(&Stdout, "", log.Ldate|log.Ltime|log.Lmicroseconds)

	for _, arg := range flag.Args() {
		i := strings.Index(arg, ":")
		if i == -1 {
			logger.Fatalf("Invalid header %q: missing ':'", arg)
		}
		k, v := arg[:i], arg[i+1:]
		v = strings.TrimSpace(v)
		headers.Add(k, v)
	}

	listener, err := net.Listen(netw, addr)
	if err != nil {
		logger.Fatalf("Error creating listener: %v", err)
	}
	defer listener.Close()

	go http.Serve(listener, http.HandlerFunc(ok))
	logger.Printf("Listening on %v", listener.Addr())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	<-sig
}
