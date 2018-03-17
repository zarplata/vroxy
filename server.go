package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/reconquest/hierr-go"
	"net/http"
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
	proxy.router.POST("/method/messages.send", proxy.handleMessagesSend)
	proxy.router.Run(addr...)
}

func (proxy *Server) handleMessagesSend(ctx *gin.Context) {
	var args MessageSendCommandArgs
	err := binding.Form.Bind(ctx.Request, &args)
	if err != nil {
		abort(
			ctx,
			http.StatusBadRequest,
			hierr.Errorf(err, "unable to bind form"),
		)
		return
	}
	proxy.commands <- VKCommand{
		AccessToken: ctx.Query("access_token"),
		Method:      "API.messages.send",
		Args:        args,
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
		"error":   err,
	})
	ctx.Abort()
}
