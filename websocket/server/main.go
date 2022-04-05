package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"go.uber.org/atomic"
	"golang.org/x/sync/errgroup"
)

var (
	upgrader = websocket.Upgrader{}
	counter  atomic.Int32
	clients  = sync.Map{}
)

type Client struct {
	conn      *websocket.Conn
	writeChan chan message
	name      string

	clients *sync.Map
	counter *atomic.Int32
	once    sync.Once
}

type message struct {
	data string
	code int
}

func NewClient(conn *websocket.Conn, counter *atomic.Int32, clients *sync.Map) *Client {
	client := &Client{
		conn:      conn,
		writeChan: make(chan message, 1),
		name:      uuid.New().String(),

		clients: clients,
		counter: counter,
		once:    sync.Once{},
	}

	conn.SetPingHandler(func(appData string) error {
		client.writeMessage("pong", websocket.PongMessage)
		return nil
	})

	counter.Inc()
	log.Println("create new client:", client.name, counter.Load())

	return client
}

func (c *Client) close() error {
	c.counter.Dec()
	log.Println("close client:", c.name, c.counter.Load())
	c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	c.clients.Delete(c.name)
	close(c.writeChan)
	return c.conn.Close()
}

func (c *Client) Close() (err error) {
	c.once.Do(func() {
		err = c.close()
	})

	return
}

func (c *Client) Read() {
	defer c.Close()
	for {
		mt, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Println("read:", err, c.name)
			return
		}

		log.Printf("recv: %d %s %s", mt, message, c.name)
	}
}

func (c *Client) writeMessage(data string, code int) {
	c.writeChan <- message{data, code}
}

func (c *Client) WriteMessage(data string) {
	c.writeChan <- message{data, websocket.TextMessage}
}

func (c *Client) Write() {
	for msg := range c.writeChan {
		err := c.conn.WriteMessage(msg.code, []byte(msg.data))
		if err != nil {
			log.Println("write:", err)
			c.Close()
			return
		}
	}
}

func (c *Client) Name() string {
	return c.name
}

func ws(w http.ResponseWriter, r *http.Request) {
	log.Print("Header", r.Header)
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	client := NewClient(c, &counter, &clients)
	clients.Store(client.Name(), client)
	go client.Read()
	go client.Write()
}

func broadcast(w http.ResponseWriter, r *http.Request) {
	clients.Range(func(key, value interface{}) bool {
		log.Println("send to client:", key)
		value.(*Client).WriteMessage("broadcast")
		return true
	})
}

func stopClients() {
	clients.Range(func(key, value interface{}) bool {
		log.Println("stop client:", key)
		value.(*Client).Close()
		return true
	})
}

func listen(ctx context.Context) error {
	srv := http.Server{
		Addr:        ":8080",
		Handler:     nil,
		BaseContext: func(net.Listener) context.Context { return ctx },
	}

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(srv.ListenAndServe)

	eg.Go(func() error {
		<-ctx.Done()

		stopClients()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			fmt.Println("Shutdown error:", err)
		}
		return err
	})

	return eg.Wait()
}

func main() {
	http.HandleFunc("/ws", ws)
	http.HandleFunc("/broadcast", broadcast)

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return listen(ctx)
	})

	fmt.Println(eg.Wait())
}
