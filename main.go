package main

import (
    "fmt"
    "log"

    "github.com/Scale3-Labs/suirpc/pkg"
)

func main() {
    client := suirpc.New("http://127.0.0.1:9000")

    serviceDiscovery, err := client.Discover()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(serviceDiscovery)
}