package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/reconquest/hierr-go"
	"net/http"
	"errors"
)

type Server struct {
	router   *gin.Engine
	commands chan<- VKCommand
}

func NewServer(commands chan<- VKCommand, verbose bool) *Server {
	if !verbose {
		gin.SetMode(gin.ReleaseMode)
	}
	return &Server{
		router:   gin.Default(),
		commands: commands,
	}
}

func (proxy *Server) Run(addr ...string) error {
	proxy.router.POST("/method/:name", proxy.handleMessagesSend)
	return proxy.router.Run(addr...)
}

func (proxy *Server) handleMessagesSend(ctx *gin.Context) {
	request := ctx.Request
	if err := request.ParseForm(); err != nil {
		abort(
			ctx,
			http.StatusBadRequest,
			hierr.Errorf(err, "unable to parse form data"),
		)
		return
	}
	request.ParseMultipartForm(32 << 10) // 32 MB

	accessToken := request.Form.Get("access_token")
	if accessToken == "" {
		accessToken = ctx.Query("access_token")
	}
	if accessToken == "" {
		abort(
			ctx,
			http.StatusBadRequest,
			errors.New("access_token is required"),
		)
		return
	}

	payload := make(map[string]interface{})
	for k, v := range request.Form {
		if v == nil || k == "access_token" {
			continue
		}
		switch len(v) {
		case 0:
			continue
		case 1:
			payload[k] = v[0]
		case 2:
			payload[k] = v
		}
	}

	proxy.commands <- VKCommand{
		AccessToken: accessToken,
		Method:      fmt.Sprintf("API.%s", ctx.Param("name")),
		Payload:     payload,
	}
	ctx.JSON(http.StatusOK, gin.H{"success": true})
}

func abort(ctx *gin.Context, code int, err error) {
	if code >= http.StatusInternalServerError {
		logger.Error(err)
	}
	ctx.Error(err)
	ctx.JSON(code, gin.H{
		"success": false,
		"error":   err.Error(),
	})
	ctx.Abort()
}
