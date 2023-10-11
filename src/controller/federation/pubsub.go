package federation

import (
	"chainsource-gateway/helpers"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/gorilla/websocket"
)

var dialer = websocket.Dialer{}
var conn *websocket.Conn = nil

func OpenWS(w http.ResponseWriter, r *http.Request) {

	var nodeUri = chi.URLParam(r, "node_uri")
	// var channelIdFromRequest = chi.URLParam(r, "channel_id")

	u := "ws://" + nodeUri + ":3050/api/v2/ws"
	log.Printf("connecting to %s", u)

	req, _ := http.NewRequest("GET", "http://"+nodeUri+":3050", nil)
	req.Header.Add("NodeID", helpers.GetNodeID())
	var err error
	conn, _, err = dialer.Dial(u, req.Header)
	if err != nil {

		var response = helpers.ChannelResultResponse{
			Success: false,
			Status:  "error connecting to websocket",
		}
		render.JSON(w, r, response)
	} else {
		var response = helpers.ChannelResultResponse{
			Success: true,
			Status:  "connected to websocket",
		}
		render.JSON(w, r, response)
	}
	// defer conn.Close()
	// done := make(chan struct{})
	go func() {
		// defer close(done)
		fmt.Printf("READING MESSAGE FROM CHANNEL")
		for {
			if conn==nil{
				break
			}
			_, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Printf("Read error %v", err)
				return
			}
			fmt.Printf("recv : %s", message)

		}
		fmt.Printf("Unsubscribed")

	}()
}

func CloseWS(w http.ResponseWriter, r *http.Request) {
	if conn != nil {
		fmt.Print("\nClosing connection\n")
		conn.Close()
		conn=nil
	}
	return

}
