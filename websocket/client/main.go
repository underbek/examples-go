package main

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
)

const (
	addr       = "localhost:8080"
	maxClients = 10_000
)

type Client struct {
	conn *websocket.Conn
}

func NewClient(addr string) (*Client, error) {
	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}

	header := http.Header{}
	header["jwt"] = []string{"my_best_jwt"}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return nil, err
	}

	client := &Client{
		conn: c,
	}

	c.SetPongHandler(func(appData string) error {
		log.Println("pong message:", appData)
		return nil
	})

	return client, nil
}

func (c *Client) Do(ctx context.Context) error {
	defer c.conn.Close()

	err := c.conn.WriteMessage(websocket.PingMessage, []byte("ping"))
	if err != nil {
		log.Println("write:", err)
		return err
	}

	time.Sleep(time.Second * 5)

	err = c.conn.WriteMessage(websocket.TextMessage, []byte("hello!!!"))
	if err != nil {
		log.Println("write:", err)
		return err
	}

	gr, ctx := errgroup.WithContext(ctx)
	gr.Go(func() error {
		<-ctx.Done()
		err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Println("write close:", err)
			return err
		}
		return ctx.Err()
	})

	gr.Go(func() error {
		for {
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return err
			}
			log.Printf("recv: %s", message)
		}
	})

	return gr.Wait()
}

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	gr, ctx := errgroup.WithContext(ctx)

	for i := 0; i < maxClients; i++ {
		client, err := NewClient(addr)
		if err != nil {
			log.Println("create client:", err)
			cancel()
			break
		}
		gr.Go(func() error {
			return client.Do(ctx)
		})
	}

	log.Fatal(gr.Wait())
}
