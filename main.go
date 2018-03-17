package main

import (
	"github.com/docopt/docopt-go"
	"github.com/kovetskiy/lorg"
	"github.com/reconquest/hierr-go"
	"strconv"
)

var version = "DEV"
var logger *lorg.Log

func main() {
	usage := `vroxy ` + version + `

Proxy server for balancing requests to VK API to avoid rate limiting.

Usage:
    vroxy (-t <token>) [options]

Options:
	-t --access-token <token>  VK access token.	
	-l --listen <address>      HTTP listen address. [default: :8080]
	-r --rps <rps>             Permissible VK API RPS. [default: 10]
	-v --verbose               Logging in debug mode.
`

	arguments, err := docopt.Parse(usage, nil, true, version, false)
	if err != nil {
		panic(err)
	}

	verbose := arguments["--verbose"].(bool)
	logger = setupLogger(verbose)

	listen := arguments["--listen"].(string)

	accessToken := arguments["--access-token"].(string)
	if err != nil {
		hierr.Fatalf(err, "unable to parse --access-token")
	}

	rps, err := strconv.Atoi(arguments["--rps"].(string))
	if err != nil {
		hierr.Fatalf(err, "unable to parse --rps")
	}

	queue := NewCommandsQueue(rps)
	queue.Run()

	vk := NewVKClient(accessToken, rps)
	vk.Run(queue.Chunks)

	server := NewServer(queue.Commands, verbose)
	server.Run(listen)
}

func setupLogger(verbose bool) *lorg.Log {

	formatTemplate := `${level} %s [${file}:${line}]`
	if verbose {
		formatTemplate = `${time:15:04:05} ${level} %s [${file}:${line}]`
	}
	newLogger := lorg.NewLog()
	newLogger.SetFormat(lorg.NewFormat(formatTemplate))
	newLogger.SetLevel(lorg.LevelInfo)

	if verbose {
		newLogger.SetLevel(lorg.LevelDebug)
	}

	return newLogger
}
