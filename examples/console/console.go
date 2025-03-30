//
// Copyright (C) 2024 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/kshard/chatter
//

package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kshard/chatter"
	"github.com/kshard/chatter/llm/bedrock"
)

var defllm = bedrock.LLAMA3_1_405B_INSTRUCT

var awsllms = []chatter.LLM{
	bedrock.LLAMA3_2_90B_INSTRUCT,
	bedrock.LLAMA3_2_11B_INSTRUCT,
	bedrock.LLAMA3_2_3B_INSTRUCT,
	bedrock.LLAMA3_2_1B_INSTRUCT,
	bedrock.LLAMA3_1_405B_INSTRUCT,
	bedrock.LLAMA3_1_70B_INSTRUCT,
	bedrock.LLAMA3_1_8B_INSTRUCT,
	bedrock.LLAMA3_0_70B_INSTRUCT,
	bedrock.LLAMA3_0_8B_INSTRUCT,
	bedrock.TITAN_TEXT_PREMIER_V1,
	bedrock.TITAN_TEXT_EXPRESS_V1,
	bedrock.TITAN_TEXT_LITE_V1,
}

// var oaillms = []chatter.LLM{
// 	openai.GPT_O1,
//
// }

type service struct {
	config []string
	routes map[string]chatter.Chatter
}

func New() (*service, error) {
	config := make([]string, 0)
	routes := make(map[string]chatter.Chatter)

	for _, llm := range awsllms {
		api, err := bedrock.New(
			bedrock.WithLLM(llm),
		)
		if err != nil {
			return nil, err
		}

		config = append(config, llm.ModelID())
		routes[llm.ModelID()] = api
		if llm.ModelID() == defllm.ModelID() {
			routes["_"] = api
		}
	}

	return &service{config: config, routes: routes}, nil
}

func (api *service) httpConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(api.config)
}

func (api *service) httpPrompt(w http.ResponseWriter, r *http.Request) {
	var prompt chatter.Prompt
	if err := json.NewDecoder(r.Body).Decode(&prompt); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var chatter chatter.Chatter
	if val, has := api.routes[r.URL.Query().Get("llm")]; has {
		chatter = val
	}

	if chatter == nil {
		chatter = api.routes["_"]
	}

	reply, err := chatter.Prompt(context.Background(), prompt.ToSeq())
	if err != nil {
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	json.NewEncoder(w).Encode(reply)
}

func main() {
	api, err := New()
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/config", api.httpConfig)
	http.HandleFunc("/prompt", api.httpPrompt)
	err = http.ListenAndServe(":8081", nil)
	if err != nil {
		panic(err)
	}
}
