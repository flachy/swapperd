package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/republicprotocol/atom-go/adapters/configs/general"
)

func main() {
	cfg, err := config.LoadConfig("../config.json")
	if err != nil {
		panic(err)
	}

	addresses := []string{}
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Authorize your ethereum address(es): ")
	for {
		text, _ := reader.ReadString('\n')
		if text == "\n" {
			break
		}
		addresses = append(addresses, strings.Trim(text, "\n"))
	}
	cfg.AuthorizedAddresses = addresses
	cfg.StoreLoc = os.Getenv("HOME") + "/.swapper/db"

	if err := cfg.Update(); err != nil {
		panic(err)
	}

}
