// Start a single instance of testsvr.
package main

import (
	"fmt"
	"os"

	"github.com/mmlt/apigw/test/server"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Printf(`Test server.
Usage: %s idp|multi <name> <ip:port>
   idp - act as an oauth2 idp mock with /oauth2/tokeninfo endpoint.
   multi - act as a multipurpose server with html and websocker endpoints.
Example: %s multi first-instance localhost:1323`, os.Args[0], os.Args[0])
		os.Exit(1)
	}
	role := os.Args[1]
	name := os.Args[2]
	port := os.Args[3]

	var svr *server.Testsvr
	switch role {
	case "idp":
		svr = server.IDP(name, port)
	case "multi":
		svr = server.Multipurpose(name, port)
	default:
		fmt.Println("unknown role", role)
		os.Exit(1)
	}

	err := svr.Run()
	fmt.Println(err)
}
