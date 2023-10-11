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
	"sync"
	"time"

	_ "github.com/gorilla/websocket"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Event handler will provide appropriate responses based on event type
// Current implementation includes server-> client AddedAsset
// Client-> Server {ping-pong, request_open_ws}
type EventHandler func(event Event, m *Manager) error

type ClientEventHandler func(event Event, c *WSClient) error

const (
	EventNotifyAssetChange = "channel_change" // emited by server
)

// TODO:ping req pong res will happen at set intervals in WSClient to ensure connection is live
// TODO: Close websocket connection as per requirement
type AssetChangeEvent struct {
	Time      time.Time `json:"time"`
	Message   string    `json:"payload"`
	ChannelId string    `json:"channelId"`
	Source    string    `json:"host"`
}

func AssetChangeHandler(event Event, m *Manager) error {

	var e AssetChangeEvent
	if err := json.Unmarshal(event.Payload, &e); err != nil {
		fmt.Printf("Error event.go line 37 %v", err)
	}
	chanId := e.ChannelId
	var e2 AssetChangeEvent
	e2.Time = e.Time

	e2.ChannelId = e.ChannelId
	e2.Message = e.Message
	e2.Source = e.Source
	data, err := json.Marshal(e2)
	if err != nil {
		fmt.Printf("Failed to marshal event %v", err)
	}

	var packet Event

	packet.Payload = data
	packet.Type = EventNotifyAssetChange
	fmt.Printf("AssetChangeHandler %s", packet.Type)
	var wg sync.WaitGroup
	for client := range m.clients {
		if _, ok := client.channels[chanId]; ok {			
			fmt.Printf("Passing packet to client")
			wg.Add(1)
			go m.SendWithWait(client, packet.Type, &wg)
			fmt.Printf("Message Sent")
		}
	}
	fmt.Printf("Exiting assetChangeHandler")
	return nil
}
