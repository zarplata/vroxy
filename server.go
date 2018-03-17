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
		router:   gin.New(),
		commands: commands,
	}
}

func (proxy *Server) Run(addr ...string) {
	proxy.router.POST("/method/:name", proxy.handleMessagesSend)
	proxy.router.Run(addr...)
}

func (proxy *Server) handleMessagesSend(ctx *gin.Context) {
	accessToken := ctx.Query("access_token")
	if accessToken == "" {
		abort(
			ctx,
			http.StatusBadRequest,
			errors.New("query `access_token` is required"),
		)
		return
	}

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
	proxy.commands <- VKCommand{
		AccessToken: accessToken,
		Method:      fmt.Sprintf("API.%s", ctx.Param("name")),
		Args:        request.Form,
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
