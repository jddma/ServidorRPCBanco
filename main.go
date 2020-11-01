package main

import (
	"./server"

	"fmt"
)

func main() {

	var port string
	fmt.Print("Digite el puerto a usar: ")
	fmt.Scanf("%s", &port)

	centralServer := new(server.Central)
	centralServer.StartServer(port)

}
