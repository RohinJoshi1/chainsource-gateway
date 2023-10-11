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
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
)

var (
	//Upgrade incoming HTTP request to persistent websocket conn
	websocketUpgrader = websocket.Upgrader{
		// CheckOrigin:     checkOrigin,
		ReadBufferSize:  2048,
		WriteBufferSize: 2048,
	}
)

var (
	UnsupportedEvent = errors.New("errunsupportedevent")
)

// func checkOrigin(r *http.Request) bool {
// 	//Check the origin of the request and return true if access is allowed
// 	origin := r.Header.Get("Origin")
// 	switch origin {
// 	case "https://localhost:7205":
// 		return true
// 	default:
// 		return true
// 	}

// }

//Manager will orchestrate between all the client websocket connections on the server
/*
	1.Holds client list per channel
	2.Mutex
	3.eventHandler

*/
type Manager struct {
	clients ClientList
	sync.RWMutex
	handlers map[string]EventHandler
	updates  chan Event
}

func NewWsManager() *Manager {
	m := &Manager{
		clients:  make(ClientList),
		handlers: make(map[string]EventHandler),
		updates:  make(chan Event),
	}
	m.setupEventHandlers()
	return m
}

func (m *Manager) setupEventHandlers() {
	m.handlers[EventNotifyAssetChange] = AssetChangeHandler
}

func (m *Manager) ServeWS(w http.ResponseWriter, r *http.Request) {
	fmt.Print("received new connection request")
	var wg sync.WaitGroup
	h := r.Header
	conn, err := websocketUpgrader.Upgrade(w, r, h)
	if err != nil {
		fmt.Print(err)
		return
	}
	client := NewWsClient(conn, m, h)
	m.addClient(client)
	m.Send(conn, "Connected to WEBSOCKET")
	wg.Add(1)
	go m.UpdateHandler(&wg)
	// go client.WriteMessage()
	// go client.ReadMessage()

	// wg.Done()

}

func (m *Manager) addClient(client *WSClient) {
	m.Lock()
	defer m.Unlock()
	m.clients[client] = true
}

func (m *Manager) RemoveClient(client *WSClient) {
	m.Lock()
	defer m.Unlock()
	fmt.Println("REMOVING CLIENT")
	if _, ok := m.clients[client]; ok {
		client.connection.Close()
		delete(m.clients, client)
	}
}

func (m *Manager) eventRouter(event Event) error {
	//When I get an event of type x , I check if there exists an appropriate handler for this event
	for {
		// event := <- m.updates
		fmt.Printf("\nEVENT ROUTER %s", event.Type)
		if handler, ok := m.handlers[event.Type]; ok {
			fmt.Printf("EVENT TYPE: %s\n", event.Type)
			if err := handler(event, m); err != nil {
				fmt.Print(err)
				return err
			}
			return nil
		} else {
			return UnsupportedEvent
		}
	}

}

func (m *Manager) UpdateHandler(wg *sync.WaitGroup) {
	KV := GetKVInstance()
	watcher, err := KV.kv.WatchAll()
	if err != nil {
		fmt.Printf("%v", err)
	}
	defer watcher.Stop()

	fmt.Printf("Watching\n")
	for {
		select {
		case v := <-watcher.Updates():
			if v == nil {
				// ignore initial value marker which is a nil KeyValueEntry
				continue
			}
			fmt.Printf("[%v][%v]: [%s]", v.Created(), v.Operation(), v.Value())
			event := createEvent(v)
			// m.updates <- event
			err := m.eventRouter(event)
			if err != nil {
				fmt.Printf("Error line 181 %v", err)
			}
		}
	}

}

func createEvent(kve nats.KeyValueEntry) Event {
	payload := struct {
		Time      time.Time `json:"time"`
		Message   string    `json:"payload"`
		ChannelId string    `json:"channelId"`
		Source    string    `json:"host"`
	}{
		Time:      kve.Created(),
		Message:   string(kve.Operation()),
		ChannelId: kve.Key(),
		Source:    GetNodeURI(),
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("%v line 113 kv.go", err)
	}

	var e Event
	//Add type based on type of event, currently supports only asset change
	e.Type = EventNotifyAssetChange
	e.Payload = payloadJSON
	return e
}
func (m *Manager) Send(conn *websocket.Conn, message string) {
	conn.WriteMessage(websocket.TextMessage, []byte(message))
}
func (m *Manager) SendWithWait(client *WSClient , message string, wg *sync.WaitGroup) {
	conn := client.connection
	err := conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err!=nil{
		m.RemoveClient(client)
	}
}
