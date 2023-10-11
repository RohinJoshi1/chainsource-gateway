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
	"os"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

type KVStore struct {
	kv nats.KeyValue
	sync.RWMutex
}

var (
	KV *KVStore
)

func NewKV() *KVStore {
	url := os.Getenv("NATS_URI")
	if url == "" {
		url = nats.DefaultURL
	}

	nc, _ := nats.Connect(url)
	js, _ := nc.JetStream()

	var kv nats.KeyValue
	var err error
	kv, err = js.CreateKeyValue(&nats.KeyValueConfig{
		Bucket: "Channels",
	})
	if err != nil {
		panic(err)
	}

	KV = &KVStore{
		kv: kv,
	}
	return KV
}

func GetKVInstance() *KVStore {
	return KV
}

func (KV *KVStore) Get(channelId string) string {
	entry, _ := KV.kv.Get(channelId)
	//will return nil if absent
	return string(entry.Value())
}

func (KV *KVStore) getLatestRevision(channelId string) uint64 {
	entry, _ := KV.kv.Get(channelId)
	return entry.Revision()
}

func (KV *KVStore) Create(channelId string, newTime string) {
	KV.Lock()
	defer KV.Unlock()
	KV.kv.Create(channelId, []byte(newTime))
	fmt.Printf("CREATE successful\n")
}

func (KV *KVStore) Put(channelId string, newTime string) {
	KV.Lock()
	defer KV.Unlock()
	KV.kv.Put(channelId, []byte(newTime))
	fmt.Printf("PUT successful\n")
}

// Get latest revision
func (KV *KVStore) Update(channelId string, newTime string) {
	KV.Lock()
	defer KV.Unlock()
	last_rev := KV.getLatestRevision(channelId)
	KV.kv.Update(channelId, []byte(newTime), uint64(last_rev))
	fmt.Printf("UPDATE successful\n")

}

func (KV *KVStore) Delete(channelId string) {
	KV.Lock()
	defer KV.Unlock()
	KV.kv.Delete(channelId)
	fmt.Printf("DELETE successful\n")
}

// This func will be called as a go routine by the websocket manager, and it emits events, these events are handled appropriately
func (KV *KVStore) Watch(m *Manager) {
	w, _ := KV.kv.WatchAll()
	defer w.Stop()
	kve := <-w.Updates()
	kve2 := kve
	//wrap this in an event struct and pass it to m.updatesChan
	//kve format :
	payload := struct {
		Time      time.Time `json:"time"`
		Message   string    `json:"payload"`
		ChannelId string    `json:"channelId"`
		Source    string    `json:"host"`
	}{
		Time:      kve2.Created(),
		Message:   string(kve2.Operation()),
		ChannelId: kve.Key(),
		Source:    GetNodeURI(),
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("%v line 113 kv.go", err)
	}

	var e Event
	//Add type based on type of event, currently supports only assetchange
	e.Type = EventNotifyAssetChange
	e.Payload = payloadJSON

	//Feed e into m.updatesChannel
	m.updates <- e
}
