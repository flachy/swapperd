package main

import (
	"errors"
	"log"
	"os"
	"path"

	"github.com/republicprotocol/renex-swapper-go/driver/config"
	"github.com/republicprotocol/renex-swapper-go/driver/keystore"
	"github.com/republicprotocol/renex-swapper-go/driver/swapper"
	"github.com/urfave/cli"
)

// Define flags for commands
var (
	networkFlag = cli.StringFlag{
		Name:  "network",
		Value: "mainnet",
		Usage: "name of the network",
	}
	keyPhraseFlag = cli.StringFlag{
		Name:  "keyphrase",
		Usage: "keyphrase to unlock the keystore file",
	}
	toFlag = cli.StringFlag{
		Name:  "to",
		Usage: "receiver address you want to withdraw the tokens to",
	}
	tokenFlag = cli.StringFlag{
		Name:  "token",
		Usage: "type of token you want to withdraw",
	}
	valueFlag = cli.Float64Flag{
		Name:  "value",
		Value: 0,
		Usage: "amount of token you want to withdraw,", // todo: specify the unit here
	}
	feeFlag = cli.Float64Flag{
		Name:  "fee",
		Value: 0,
		Usage: "amount of fee you want to pay for the withdraw transaction,", // todo: specify the unit here
	}
	portFlag = cli.Int64Flag{
		Name:  "port",
		Value: 18516,
		Usage: "port on which the http server runs,",
	}
)

func main() {
	app := cli.NewApp()

	// Define sub-commands
	app.Commands = []cli.Command{
		{
			Name:  "http",
			Usage: "start running the swapper ",
			Flags: []cli.Flag{networkFlag, keyPhraseFlag, portFlag},
			Action: func(c *cli.Context) {
				swapper, err := initializeSwapper(c)
				if err != nil {
					panic(err)
				}
				swapper.Http(c.Int64("port"))
			},
		},
		{
			Name:  "withdraw",
			Usage: "withdraw the funds in the swapper accounts",
			Action: func(c *cli.Context) error {
				swapper, err := initializeSwapper(c)
				if err != nil {
					return err
				}

				return withdraw(c, swapper)
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func initializeSwapper(ctx *cli.Context) (swapper.Swapper, error) {
	network := ctx.String("network")
	keyPhrase := ctx.String("keyphrase")

	cfg, err := config.New(path.Join(os.Getenv("HOME"), ".swapper"), network)
	if err != nil {
		return nil, err
	}

	ks, err := keystore.LoadFromFile(cfg, keyPhrase)
	if err != nil {
		return nil, err
	}

	return swapper.NewSwapper(cfg, ks), nil
}

func withdraw(ctx *cli.Context, swapper swapper.Swapper) error {
	receiver := ctx.String("to")
	if receiver == "" {
		return errors.New("receiver address cannot be empty")
	}
	token := ctx.String("token")
	if token == "" {
		return errors.New("please enter a valid withdraw token")
	}
	return swapper.Withdraw(token, receiver, ctx.Float64("value"), ctx.Float64("fee"))
}
