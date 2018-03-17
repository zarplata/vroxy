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
	AccessToken string
	client      *http.Client
	version     string
}

type Command struct {
	Method string
	Args   interface{}
}

type Commands []Command

type MessageSendCommandArgs struct {
	UserID  int    `json:"user_id" form:"user_id" binding:"required"`
	Message string `json:"message" form:"message" binding:"required"`
}

type ExecuteResponsePayload struct {
	Response []interface{} `json:"response"`
}

func NewVKClient(accessToken string, rps int) *VKClient {
	client := &http.Client{Transport: &http.Transport{
		MaxIdleConnsPerHost: rps * 2,
		MaxIdleConns:        rps * 2,
	}}

	return &VKClient{
		AccessToken: accessToken,
		client:      client,
		version:     "5.73",
	}
}

func (vk *VKClient) Run(chunks <-chan Commands) {
	go func() {
		for {
			select {
			case commands := <-chunks:
				go func(commands Commands) {
					err, cnt := vk.execute(commands)
					if err != nil {
						e := hierr.Errorf(
							err,
							"can't execute commands: %+v",
							commands,
						)
						if cnt > 0 {
							logger.Warning(e)
						} else {
							logger.Error(e)
						}
					}
				}(commands)
			}
		}
	}()
}

func (vk *VKClient) execute(commands Commands) (error, int) {
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
	query.Add("access_token", vk.AccessToken)
	query.Add("version", vk.Version)
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

	payload := ExecuteResponsePayload{}
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

func compileCode(commands []Command) (string, error) {
	commandsCode := make([]string, len(commands))
	for i, command := range commands {
		b, err := json.Marshal(command.Args)
		if err != nil {
			return "", hierr.Errorf(
				err,
				"can't marshal command args %+v",
				command.Args,
			)
		}
		commandsCode[i] = fmt.Sprintf("%s(%s)", command.Method, string(b))
	}
	code := fmt.Sprintf("return [%s];", strings.Join(commandsCode, ","))
	return code, nil
}

func (payload *ExecuteResponsePayload) getFailed() int {
	cnt := 0
	for _, id := range payload.Response {
		switch id.(type) {
		case bool:
			cnt++
		}
	}
	return cnt
}
