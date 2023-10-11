/*
 * Copyright 2023 Unisys Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package helpers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
)

type ClientList map[*WSClient]bool
type void struct{}

var member void

type WSClient struct {
	connection *websocket.Conn
	manager    *Manager
	//Hold the list of channels client is listening for
	channels map[string]void
	//ingress to get updates fed
	ingress chan Event
}

var (
	//Wait 10 seconds for pong response from server
	pongWait = 10 * time.Second
	//Send ping message at fixed interval < pongWait ,
	//to ensure new ping msg is not sent before pong res (90% PONG_WAIT)
	pingInterval = 9 * pongWait / 10
)

func NewWsClient(conn *websocket.Conn, manager *Manager, h http.Header) *WSClient {
	fmt.Printf("Client created\n")
	nodeID := h.Get("NodeID")
	ch := getChannels(nodeID)
	fmt.Printf("HEADER nodeID %s\n", nodeID)
	return &WSClient{
		connection: conn,
		manager:    manager,
		channels:   ch,
	}
}

func getChannels(nodeID string) map[string]void {
	chans := make(map[string]void)
	details := GetNodeDetailsFromId(GetNodeURI())

	//From node connection get channel connection
	node_conns := details.NodeConnections
	channels := make([]ChannelConnection, 0)
	for _, nodes := range node_conns {
		if nodes.NodeId == nodeID {
			channels = append(channels, nodes.ChannelConnections...)
			break
		}
	}

	for _, cc := range channels {
		chans[cc.ChannelId] = member
		fmt.Printf("Channel found %s\n", cc.ChannelId)
	}

	return chans

}
func (c *WSClient) Unsubscribe() {
	defer c.connection.Close();
	c.manager.RemoveClient(c);
}

func GetNodeDetailsFromId(nodeId string) Node {
	nc, ncErr := nats.Connect(os.Getenv("NATS_URI"))
	if ncErr != nil {
		log.Err(ncErr).Msgf(NatsConnectError)
		return Node{}
	}
	defer nc.Close()

	requestMarshal, marshalErr := json.Marshal(nodeId)
	if marshalErr != nil {
		log.Err(marshalErr).Msgf(MarshalErr)
		return Node{}
	}

	// Send the request
	msg, msgErr := nc.Request("node.details", requestMarshal, TimeOut*time.Second)
	if msgErr != nil {
		log.Err(msgErr).Msgf(MsgErr)
		return Node{}
	}

	var response NodeResultResponse
	unmarshalErr := json.Unmarshal([]byte(string(msg.Data)), &response)
	if unmarshalErr != nil {
		log.Err(unmarshalErr).Msgf(UnmarshalErr)
		return Node{}
	}

	return response.Result[0]
}



/** 
	NOTE: EventListener, ReadMessage and WriteMessage are currently not being used in the primary flow 
	and has been replaced by send message in ws_manager, reading from connection is done by client in pub_sub
*/
func (c *WSClient) EventListener() {
	fmt.Printf("Inside update EventListener\n")
	for {
		event := <-c.ingress
		fmt.Printf("EVENTS %v", event.Type)
		var e AssetChangeEvent
		if err := json.Unmarshal(event.Payload, &e); err != nil {
			fmt.Printf("error event.go line 37 %v", err)
		}
		chanId := e.ChannelId

		s := fmt.Sprintf("asset updated on %s", chanId)
		switch event.Type {
		case EventNotifyAssetChange:
			log.Printf("%s", s) //-> CLient KV, correspon node/channel -> updates

		default:
			fmt.Print("ingress channel empty or invalid type\n")
			return
		}
	}
}

func (c *WSClient) WriteMessage() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		// c.manager.RemoveClient(c)
	}()

	for {
		select {
		case <-ticker.C:
			if err := c.connection.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Printf("Err %v", err)
				return
			}
		case event, ok := <-c.ingress:
			if !ok {
				continue
			}
			fmt.Printf("Client event, %s",event.Type)
			data, err := json.Marshal(event)
			fmt.Printf("Writing this to client %s", string(data))
			if err != nil {
				fmt.Printf("err 167 ws_client %v", err)
				return
			}
			if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
				fmt.Printf("Err line 171 ws_client.go %v", err)
			}
			fmt.Printf("Message Sent")
		default:
			return

		}
	}

}

func (c *WSClient) ReadMessage() {
	//ReadMessage is a go routine function
	defer func() {
		// c.manager.RemoveClient(c)
	}()
	

	if err := c.connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Printf("Err %v", err)
		return
	}
	c.connection.SetPongHandler(c.connection.PongHandler())

	for {
		//As long as there are messages, keep socket open
		_, msg, err := c.connection.ReadMessage()
		if err != nil {
			//TO DO
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Print(err)
			}
			break
		}
		//Unmarshal received json into an event
		var req Event
		if err := json.Unmarshal(msg, &req); err != nil {
			log.Printf("Error line 65 ws client %v", err)
			break
		}
		fmt.Printf(string(req.Payload))

	}

}
func (c *WSClient) PongHandler(msg string) error {
	return c.connection.SetReadDeadline(time.Now().Add(pongWait))
}
