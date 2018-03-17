package main

import (
	"time"
)

type CommandQueue struct {
	RPS       int
	ChunkSize int
	Commands  chan Command
	Chunks    chan Commands
}

func NewCommandsQueue(rps int) *CommandQueue {
	return &CommandQueue{
		RPS:       rps,
		ChunkSize: 25,
		Commands:  make(chan Command),
		Chunks:    make(chan Commands),
	}
}

func (queue *CommandQueue) Run() {
	go func() {
		commands := make(Commands, 0)
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case command := <-queue.Commands:
				logger.Debugf("append command to queue: %+v", command)
				commands = append(commands, command)

			case <-ticker.C:
				total := len(commands)
				if total == 0 {
					continue
				}
				delivered := 0
				i := 0
				for ; i < queue.RPS && delivered < total; i++ {
					size := len(commands)
					if size > queue.ChunkSize {
						size = queue.ChunkSize
					}
					queue.Chunks <- commands[:size]

					commands = commands[size:]
					delivered += size
				}
				logger.Infof(
					"delivered %d of %d commands in %d batches",
					delivered,
					total,
					i,
				)
			}
		}
	}()
}
