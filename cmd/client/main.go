package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"

	"github.com/dobegor/chat/common"
	"github.com/gorilla/websocket"
)

var (
	serverAddr string
)

func init() {
	flag.StringVar(&serverAddr, "s", "localhost:9090", "server address")
}

func read() string {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return text
}

func main() {
	flag.Parse()
	fmt.Print("Enter your nickname: ")
	var username string
	fmt.Scanln(&username)

	u := url.URL{
		Scheme: "ws",
		Host:   serverAddr,
		Path:   "/chat/" + username,
	}

	fmt.Println("Connecting chat client to: ", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)

	if err != nil {
		fmt.Println("Oops, an error occured: ", err)
		return
	}

	defer c.Close()

	go func() {
		for {
			_, rawMsg, err := c.ReadMessage()

			if err != nil {
				fmt.Println("Houston, we've got a problem: ", err)
				os.Exit(1)
			}

			msg, err := common.Decode(rawMsg)

			if err != nil {
				fmt.Println("I don't know what I've just got: ", err)
				os.Exit(1)
			}

			fmt.Printf("%s: %s\n", msg.Username, msg.Content)
		}

	}()

	for {
		message := read()
		err := c.WriteMessage(websocket.BinaryMessage, common.Encode(&common.Message{message, ""}))

		if err != nil {
			fmt.Println("Oops, an error occured: ", err)
			return
		}
	}

}
