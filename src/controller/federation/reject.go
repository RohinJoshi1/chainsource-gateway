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

package federation

import (
	"chainsource-gateway/helpers"
	"chainsource-gateway/responses"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/nats-io/nats.go"
)

// RejectRequest is a controller function to reject the federation request
func RejectRequest(w http.ResponseWriter, r *http.Request) {
	logger.Info().Msgf("Received request to reject federation request")

	// Connect to NATS
	nc, ncErr := nats.Connect(os.Getenv("NATS_URI"))
	if ncErr != nil {
		logger.Err(ncErr).Msgf(helpers.NatsConnectError)
		helpers.HandleError(w, r, helpers.NatsConnectError)
		return
	}
	defer nc.Close()

	// Get body contents
	var rejectRequest helpers.FederationRequestOperations
	json.NewDecoder(r.Body).Decode(&rejectRequest)

	request := helpers.FederationRequestBody{
		RequestID: chi.URLParam(r, "request_id"),
		Type:      rejectRequest.Type,
	}

	requestMarshal, marshalErr := json.Marshal(request)
	if marshalErr != nil {
		logger.Err(marshalErr).Msgf(helpers.MarshalErr)
		helpers.HandleError(w, r, helpers.MarshalErr)
		return
	}

	// Send the request
	msgReject, msgRejectErr := nc.Request("federation.reject", requestMarshal, helpers.TimeOut*time.Second)
	if msgRejectErr != nil {
		logger.Err(msgRejectErr).Msgf(helpers.MsgErr)
		helpers.HandleError(w, r, helpers.MsgErr)
		return
	}
	logger.Info().Msgf(helpers.Response+" %s\n", msgReject.Data)

	var msgRejectResponse helpers.FederationResultResponse
	unmarshalErr := json.Unmarshal([]byte(string(msgReject.Data)), &msgRejectResponse)
	if unmarshalErr != nil {
		logger.Err(unmarshalErr).Msgf(helpers.UnmarshalErr)
		helpers.HandleError(w, r, helpers.UnmarshalErr)
		return
	}

	if msgRejectResponse.Success {
		msg, msgErr := nc.Request("federation.one", requestMarshal, helpers.TimeOut*time.Second)
		if msgErr != nil {
			logger.Err(msgErr).Msgf(helpers.MsgErr)
			helpers.HandleError(w, r, helpers.MsgErr)
			return
		}

		logger.Info().Msgf(helpers.Response+" %s\n", msg.Data)

		// Use the response
		var response helpers.FederationResultResponse
		unmarshalResErr := json.Unmarshal([]byte(string(msg.Data)), &response)
		if unmarshalResErr != nil {
			logger.Err(unmarshalResErr).Msgf(helpers.UnmarshalErr)
			helpers.HandleError(w, r, helpers.UnmarshalErr)
			return
		}
		rejectRequestSelf := helpers.FederationRequestOperations{
			Type:      "REJECT",
			NodeID:    helpers.GetNodeID(),
			ChannelID: response.Result[0].ChannelID,
		}
		marshalRequest, err := json.Marshal(rejectRequestSelf)
		if err != nil {
			logger.Err(marshalErr).Msgf(helpers.MarshalErr)
			helpers.HandleError(w, r, helpers.MarshalErr)
			return
		}
		var host = response.Result[0].NodeURI + ":7205"
		updateRes, updateErr := helpers.PostJSONRequest("https://"+host+"/api/v2/federation/requests/nodes/update", []byte(marshalRequest))
		if updateErr != nil {
			http.Error(w, helpers.UpdateErr, http.StatusInternalServerError)
			helpers.HandleError(w, r, helpers.UpdateErr)
			return
		}
		data, err := io.ReadAll(updateRes.Body)
		if err != nil {
			http.Error(w, helpers.ReadingErr, http.StatusInternalServerError)
			helpers.HandleError(w, r, helpers.ReadingErr)
			return
		}
		parsedData, err := helpers.ParseJSONData(data)
		logger.Info().Msgf("Parsed Data %v", parsedData)

		if err != nil {
			http.Error(w, helpers.ParseErr, http.StatusInternalServerError)
			helpers.HandleError(w, r, helpers.ParseErr)
			return
		}

		if msgRejectResponse.Success {
			render.Render(w, r, responses.SuccessfulOkResponse(msgRejectResponse.Status))
		} else {
			render.Render(w, r, responses.ErrCustom(errors.New(msgRejectResponse.Status)))
		}
	} else {
		render.Render(w, r, responses.ErrCustom(errors.New(msgRejectResponse.Status)))
	}
}
