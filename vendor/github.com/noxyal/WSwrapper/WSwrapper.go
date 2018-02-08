package WSwrapper

import (
	"github.com/gorilla/websocket"
	"time"
	"net"
	"net/url"
	"net/http"
	"io"
)

type WSconn struct{ 
	*websocket.Conn
	r io.Reader
}
func (ws WSconn) Read(b []byte) (int, error) {
	var n int
	var err error
	if ws.r != nil {
		n, err = ws.r.Read(b)
		if n < len(b) {
			ws.r = nil
		}
	} else {
		_, r, err := ws.NextReader()
		if err != nil {
			return 0, err
		}
		n, err = r.Read(b)
		if n < len(b) {
			ws.r = nil
		} else {
			ws.r = r
		}
	}
	return n, err
}
func (ws WSconn) Write(b []byte) (int, error) {
	err := ws.WriteMessage(websocket.TextMessage, b)
	return len(b), err
}
func (ws WSconn) SetDeadline(t time.Time) (error) {
	err := ws.Conn.SetReadDeadline(t)
	if err != nil {
		return err
	}
	err = ws.Conn.SetWriteDeadline(t)
	if err != nil {
		return err
	}
	return nil
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}
func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

type Dialer struct {}

func (d *Dialer) Dial(network, address string) (net.Conn, error){
	u := url.URL{Scheme: "ws", Host: address, Path: "/"}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	return WSconn{conn, nil}, err
}
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}
func Serve(ln net.Listener, address string, handler func(net.Conn) ) {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil { 
			println(err)
			panic(err.Error())
		}
		handler(WSconn{conn, nil})
	})
	server := &http.Server{Addr: address}
	e := server.Serve(tcpKeepAliveListener{ln.(*net.TCPListener)})
	if e != nil {
		panic(e)
	}
}