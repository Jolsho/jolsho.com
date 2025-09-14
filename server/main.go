package main

import "server/servers"


func main() {
	state := servers.NewServerState()
	start_https := servers.BuildHttps(state)
	start_nginx := servers.BuildNginxServer(state)

	go start_https()
	start_nginx()
}


