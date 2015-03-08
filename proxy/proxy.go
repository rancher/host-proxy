package proxy

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"syscall"

	log "github.com/Sirupsen/logrus"

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

	conn, err := upgrader.Upgrade(resp, req, nil)
	if err != nil {
		return err
	}

	log.Infof("%s : starting connection", req.RemoteAddr)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	cleanup := func() {
		conn.Close()
		clientConn.Close()
		wg.Done()
	}

	go func() {
		defer cleanup()

		err := copyData(conn, clientConn)
		if err != nil {
			log.Infof("%s : error in output stream : %v", req.RemoteAddr, err)
		}
	}()

	wg.Add(1)
	go func() {
		defer cleanup()

		err := copyData(clientConn, conn)
		if err != nil {
			log.Infof("%s : error in input stream : %v", req.RemoteAddr, err)
		}
	}()

	wg.Wait()

	log.Infof("%s : closing connection", req.RemoteAddr)
	return nil
}

func copyData(dst *websocket.Conn, src *websocket.Conn) error {
	for {
		messageType, bytes, err := src.ReadMessage()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		err = dst.WriteMessage(messageType, bytes)
		if netErr, ok := err.(*net.OpError); ok && netErr.Err == syscall.EPIPE {
			// This is a broken pipe error which we can safely ignore
			return nil
		} else if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
	}

	return nil
}

func init() {
	upgrader.CheckOrigin = func(r *http.Request) bool {
		// allow all connections by default
		return true
	}
}
