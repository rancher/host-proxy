package proxy

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/rancherio/host-api/auth"
)

const (
	IP_ADDRESS string = "ipAddress"
	PORT       string = "port"
	TOKEN      string = "token"
	URL        string = "url"
)

var (
	dialer   = &websocket.Dialer{}
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func getUrl(req *http.Request) string {
	vars := mux.Vars(req)
	token := auth.GetToken(req)
	tokenString := req.URL.Query().Get("token")

	if token == nil {
		return ""
	}

	host, ok := token.Claims[IP_ADDRESS]
	if !ok {
		return ""
	}

	port, ok := token.Claims[PORT]
	if !ok {
		return ""
	}

	return fmt.Sprintf("ws://%s:%v/%s?token=%s", host, port, vars[URL], tokenString)
}

func Serve(resp http.ResponseWriter, req *http.Request) error {
	url := getUrl(req)
	if url == "" {
		return nil
	}

	clientConn, _, err := dialer.Dial(url, http.Header{})
	if err != nil {
		return err
	}

	defer clientConn.Close()

	conn, err := upgrader.Upgrade(resp, req, nil)
	if err != nil {
		return err
	}

	for {
		messageType, bytes, err := clientConn.ReadMessage()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		err = conn.WriteMessage(messageType, bytes)
		if netErr, ok := err.(*net.OpError); ok && netErr.Err == syscall.EPIPE {
			// This is a broken pipe error which we can safely ignore
			return nil
		} else if err != nil {
			return err
		}
	}

	return err
}

func init() {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		// allow all connections by default
		return true
	}
}
