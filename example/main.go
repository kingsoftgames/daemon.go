// main
package main

import (
	"fmt"
	"os"

	"daemon.go"
	"daemon.go/example/service"
)

func main() {
	dservice := &daemon.Service{}

	dservice.DaemonService = &service.DaemonService{}

	//ret, err := service.Manage(dservice, daemon.Mode_Product)
	ret, err := dservice.Standlaone("zone1", 9988)
	if err != nil {
		fmt.Println(ret, "\nError: ", err)
		os.Exit(1)
	}
	fmt.Println(ret)
}
