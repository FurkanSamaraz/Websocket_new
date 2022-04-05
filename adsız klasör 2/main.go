package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"

	ws "github.com/aopoltorzhicky/go_kraken/websocket"
)

var addr = flag.String("addr", ":8000", "http service address")
var askPrc string

func converter(con json.Number) (s string) {
	f, _ := con.Float64()

	askPrice_3 := f * 20
	s = fmt.Sprintf("%f", askPrice_3)
	return s
}

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	kraken := ws.NewKraken(ws.ProdBaseURL)
	if err := kraken.Connect(); err != nil {
		log.Fatalf("Error connecting to web socket: %s", err.Error())
	}

	// subscribe to BTCUSD`s book
	if err := kraken.SubscribeBook([]string{ws.BTCUSD}, ws.Depth10); err != nil {
		log.Fatalf("SubscribeTicker error: %s", err.Error())
	}

	// subscribe to BTCUSD`s candles
	if err := kraken.SubscribeCandles([]string{ws.BTCUSD}, ws.Interval1440); err != nil {
		log.Fatalf("SubscribeTicker error: %s", err.Error())
	}

	// subscribe to BTCUSD`s ticker
	if err := kraken.SubscribeTicker([]string{ws.BTCUSD}); err != nil {
		log.Fatalf("SubscribeTicker error: %s", err.Error())
	}

	// subscribe to BTCUSD`s trades
	if err := kraken.SubscribeTrades([]string{ws.BTCUSD}); err != nil {
		log.Fatalf("SubscribeTicker error: %s", err.Error())
	}

	// subscribe to BTCUSD`s spread
	if err := kraken.SubscribeSpread([]string{ws.BTCUSD}); err != nil {
		log.Fatalf("SubscribeTicker error: %s", err.Error())
	}

	for {
		select {
		case <-signals:
			log.Warn("Stopping...")
			if err := kraken.Close(); err != nil {
				log.Fatal(err)
			}
			return
		case update := <-kraken.Listen():
			switch data := update.Data.(type) {
			case []ws.Trade:
				log.Printf("----Trades of %s----", update.Pair)
				for i := range data {
					log.Printf("Price: %s", data[i].Price.String())
					log.Printf("Volume: %s", data[i].Volume.String())
					log.Printf("Time: %s", data[i].Time.String())
					log.Printf("Order type: %s", data[i].OrderType)
					log.Printf("Side: %s", data[i].Side)
					log.Printf("Misc: %s", data[i].Misc)
					fmt.Println("===================================================")
				}
			case ws.OrderBookUpdate:

				log.Printf("----Aksk of %s----", update.Pair)
				for _, ask := range data.Asks {
					askPrc = converter(ask.Price)
					var askVlm = converter(ask.Volume)
					//fmt.Println(askStr)
					log.Printf("%s | %s", askPrc, askVlm)

				}

				done = make(chan interface{})    // Channel to indicate that the receiverHandler is done
				interrupt = make(chan os.Signal) // Channel to listen for interrupt signal to terminate gracefully

				signal.Notify(interrupt, os.Interrupt) // Notify the interrupt channel for SIGINT

				socketUrl := "ws://localhost:8080" + "/socket"
				conn, _, err := websocket.DefaultDialer.Dial(socketUrl, nil)
				if err != nil {
					log.Fatal("Error connecting to Websocket Server:", err)
				}
				defer conn.Close()
				go receiveHandler(conn)

				// Our main loop for the client
				// We send our relevant packets here
				for {
					select {
					case <-time.After(time.Duration(1) * time.Millisecond * 1000):
						// Send an echo packet every second
						err := conn.WriteMessage(websocket.TextMessage, []byte(askPrc))
						if err != nil {
							log.Println("Error during writing to websocket:", err)
							return
						}

					case <-interrupt:
						// We received a SIGINT (Ctrl + C). Terminate gracefully…
						log.Println("Received SIGINT interrupt signal. Closing all pending connections")

						// Close our websocket connection
						err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
						if err != nil {
							log.Println("Error during closing websocket:", err)
							return
						}

						select {
						case <-done:
							log.Println("Receiver Channel Closed! Exiting….")
						case <-time.After(time.Duration(1) * time.Second):
							log.Println("Timeout in closing receiving channel. Exiting….")
						}
						return
					}
				}
				fmt.Println("===================================================")
				log.Printf("----Bids of %s----", update.Pair)
				for _, bid := range data.Bids {
					var bidPrc = converter(bid.Price)
					var bidVlm = converter(bid.Volume)
					log.Printf("%s | %s", bidPrc, bidVlm)
				}
			default:
			}
		}
	}
}

var done chan interface{}
var interrupt chan os.Signal

func receiveHandler(connection *websocket.Conn) {
	defer close(done)
	for {
		_, msg, err := connection.ReadMessage()
		if err != nil {
			log.Println("Error in receive:", err)
			return
		}
		log.Printf("Received: %s\n", msg)
	}
}

/*package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/webdeveloppro/golang-websocket-client/pkg/client"

	ws "github.com/aopoltorzhicky/go_kraken/websocket"
)

var addr = flag.String("addr", ":8000", "http service address")
var askPrc string

func converter(con json.Number) (s string) {
	f, _ := con.Float64()

	askPrice_3 := f * 20
	s = fmt.Sprintf("%f", askPrice_3)
	return s
}

func main() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	kraken := ws.NewKraken(ws.ProdBaseURL)
	if err := kraken.Connect(); err != nil {
		log.Fatalf("Error connecting to web socket: %s", err.Error())
	}

	// subscribe to BTCUSD`s book
	if err := kraken.SubscribeBook([]string{ws.BTCUSD}, ws.Depth10); err != nil {
		log.Fatalf("SubscribeTicker error: %s", err.Error())
	}

	// subscribe to BTCUSD`s candles
	if err := kraken.SubscribeCandles([]string{ws.BTCUSD}, ws.Interval1440); err != nil {
		log.Fatalf("SubscribeTicker error: %s", err.Error())
	}

	// subscribe to BTCUSD`s ticker
	if err := kraken.SubscribeTicker([]string{ws.BTCUSD}); err != nil {
		log.Fatalf("SubscribeTicker error: %s", err.Error())
	}

	// subscribe to BTCUSD`s trades
	if err := kraken.SubscribeTrades([]string{ws.BTCUSD}); err != nil {
		log.Fatalf("SubscribeTicker error: %s", err.Error())
	}

	// subscribe to BTCUSD`s spread
	if err := kraken.SubscribeSpread([]string{ws.BTCUSD}); err != nil {
		log.Fatalf("SubscribeTicker error: %s", err.Error())
	}

	for {
		select {
		case <-signals:
			log.Warn("Stopping...")
			if err := kraken.Close(); err != nil {
				log.Fatal(err)
			}
			return
		case update := <-kraken.Listen():
			switch data := update.Data.(type) {
			case []ws.Trade:
				log.Printf("----Trades of %s----", update.Pair)
				for i := range data {
					log.Printf("Price: %s", data[i].Price.String())
					log.Printf("Volume: %s", data[i].Volume.String())
					log.Printf("Time: %s", data[i].Time.String())
					log.Printf("Order type: %s", data[i].OrderType)
					log.Printf("Side: %s", data[i].Side)
					log.Printf("Misc: %s", data[i].Misc)
					fmt.Println("===================================================")
				}
			case ws.OrderBookUpdate:

				log.Printf("----Aksk of %s----", update.Pair)
				for _, ask := range data.Asks {
					askPrc = converter(ask.Price)
					var askVlm = converter(ask.Volume)
					//fmt.Println(askStr)
					log.Printf("%s | %s", askPrc, askVlm)

				}
				flag.Parse()

				client, err := client.NewWebSocketClient(*addr, "frontend")
				if err != nil {
					panic(err)
				}
				fmt.Println("Connecting")

				go func() {
					// write down data every 100 ms
					ticker := time.NewTicker(time.Millisecond * 1500)
					i := 0
					for range ticker.C {
						err := client.Write(askPrc)
						if err != nil {
							fmt.Printf("error: %v, writing error\n", err)
						}
						i++
					}
				}()

				// Close connection correctly on exit
				sigs := make(chan os.Signal, 1)

				// `signal.Notify` registers the given channel to
				// receive notifications of the specified signals.
				signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

				// The program will wait here until it gets the
				<-sigs
				client.Stop()
				fmt.Println("Goodbye")
				fmt.Println("===================================================")
				log.Printf("----Bids of %s----", update.Pair)
				for _, bid := range data.Bids {
					var bidPrc = converter(bid.Price)
					var bidVlm = converter(bid.Volume)
					log.Printf("%s | %s", bidPrc, bidVlm)
				}
			default:
			}
		}
	}
}
*/
