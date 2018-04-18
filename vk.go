package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/reconquest/hierr-go"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type VKClient struct {
	client  *http.Client
	version string
}

type VKCommand struct {
	AccessToken string
	Method      string
	Payload     interface{}
}

type VKCommands []VKCommand

type VKCommandsChunk struct {
	AccessToken string
	Commands    []VKCommand
}

type VKExecuteResponsePayload struct {
	Response []interface{} `json:"response"`
}

func NewVKClient(rps int, version string) *VKClient {
	client := &http.Client{Transport: &http.Transport{
		MaxIdleConnsPerHost: rps * 2,
		MaxIdleConns:        rps * 2,
	}}

	return &VKClient{
		client:  client,
		version: version,
	}
}

func (vk *VKClient) Run(chunksCh <-chan VKCommandsChunk) {
	go func() {
		for {
			select {
			case chunk := <-chunksCh:
				go func(chunk VKCommandsChunk) {
					err, cnt := vk.execute(chunk.AccessToken, chunk.Commands)
					if err != nil {
						e := hierr.Errorf(
							err,
							"can't execute chunk of commands: %+v",
							chunk,
						)
						if cnt > 0 {
							logger.Warning(e)
						} else {
							logger.Error(e)
						}
					}
				}(chunk)
			}
		}
	}()
}

func (vk *VKClient) execute(
	accessToken string,
	commands VKCommands,
) (error, int) {
	total := len(commands)

	logger.Debugf("sending execute request with %d commands", total)

	code, err := compileCode(commands)
	if err != nil {
		return hierr.Errorf(err, "unable to compile code"), 0
	}

	form := url.Values{"code": []string{code}}
	request, err := http.NewRequest(
		"POST",
		"https://api.vk.com/method/execute",
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return hierr.Errorf(err, "can't create http request"), 0
	}

	query := request.URL.Query()
	query.Add("access_token", accessToken)
	query.Add("version", vk.version)
	request.URL.RawQuery = query.Encode()

	response, err := vk.client.Do(request)
	if err != nil {
		return hierr.Errorf(err, "can't do http request"), 0
	}
	defer response.Body.Close()

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return hierr.Errorf(err, "can't read response body"), 0
	}

	payload := VKExecuteResponsePayload{}
	err = json.Unmarshal([]byte(content), &payload)
	if err != nil {
		return hierr.Errorf(err, "can't json unmarshal response body"), 0
	}

	failed := payload.getFailed()
	var e error
	if failed > 0 {
		e = errors.New(fmt.Sprintf(
			"failed to deliver %d/%d messages: %s",
			failed,
			total,
			content,
		))
	}
	return e, total - failed
}

func compileCode(commands []VKCommand) (string, error) {
	commandsCode := make([]string, len(commands))
	for i, command := range commands {
		b, err := json.Marshal(command.Payload)
		if err != nil {
			return "", hierr.Errorf(
				err,
				"can't marshal command payload %+v",
				command.Payload,
			)
		}
		commandsCode[i] = fmt.Sprintf("%s(%s)", command.Method, string(b))
	}
	code := fmt.Sprintf("return [%s];", strings.Join(commandsCode, ","))
	return code, nil
}

func (payload *VKExecuteResponsePayload) getFailed() int {
	cnt := 0
	for _, id := range payload.Response {
		switch id.(type) {
		case bool:
			cnt++
		}
	}
	return cnt
}
