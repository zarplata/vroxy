package main

import (
	"time"
)

type CommandQueue struct {
	RPS        int
	ChunkSize  int
	CommandsCh chan VKCommand
	ChunksCh   chan VKCommandsChunk
}

func NewCommandsQueue(rps int) *CommandQueue {
	return &CommandQueue{
		RPS:        rps,
		ChunkSize:  25,
		CommandsCh: make(chan VKCommand),
		ChunksCh:   make(chan VKCommandsChunk),
	}
}

func (queue *CommandQueue) Run() {
	go func() {
		buffer := make(map[string]VKCommands)
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case command := <-queue.CommandsCh:
				logger.Debugf("append command to queue: %+v", command)
				buffer[command.AccessToken] = append(
					buffer[command.AccessToken],
					command,
				)

			case <-ticker.C:
				for accessToken, commands := range buffer {
					logger.Debugf(
						"delivering %d commands for access token %s",
						len(commands),
						accessToken,
					)
					if queue.deliver(commands, accessToken) {
						delete(buffer, accessToken)
					}
				}
			}
		}
	}()
}

func (queue *CommandQueue) deliver(
	commands VKCommands,
	accessToken string,
) bool {
	total := len(commands)
	if total == 0 {
		return true
	}
	delivered := 0
	i := 0
	for ; i < queue.RPS && delivered < total; i++ {
		size := len(commands)
		if size > queue.ChunkSize {
			size = queue.ChunkSize
		}
		queue.ChunksCh <- VKCommandsChunk{
			AccessToken: accessToken,
			Commands: commands[:size],
		}
		commands = commands[size:]
		delivered += size
	}
	logger.Infof(
		"delivered %d of %d commands in %d batches",
		delivered,
		total,
		i,
	)
	return total == delivered
}