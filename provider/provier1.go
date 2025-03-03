package main

import (
	imp2 "zRPC/provider/imp"
	"zRPC/server/imp"
)

func main() {
	registry := imp.NewRegistry()

	registry.LocalRegistry("IUserService", &imp2.UserService{Id: 1, Name: "chu", Email: "xxx"})
	server := imp.NewServer()
	server.SetRegistry(registry)
	server.Start()
}
