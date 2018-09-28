package renex

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/republicprotocol/renex-swapper-go/adapter/config"
	"github.com/republicprotocol/renex-swapper-go/domain/swap"
	"github.com/republicprotocol/renex-swapper-go/domain/token"
	"github.com/republicprotocol/renex-swapper-go/service/logger"
)

var (
	ErrVerificationFailed = fmt.Errorf("Given order id does not exist or belong to an authorized trader")
)

type Binder interface {
	GetOrderMatch(orderID [32]byte, waitTill int64) (swap.Match, error)
}

type binder struct {
	logger.Logger
	config.Config
	*Orderbook
	*RenExSettlement
}

func NewBinder(conf config.Config, logger logger.Logger) (Binder, error) {
	conn, err := NewConnWithConfig(conf)
	if err != nil {
		return nil, err
	}

	settlement, err := NewRenExSettlement(conn.RenExSettlement, bind.ContractBackend(conn.Client))
	if err != nil {
		return nil, err
	}

	orderbook, err := NewOrderbook(common.HexToAddress(conf.RenEx.Orderbook), bind.ContractBackend(conn.Client))
	if err != nil {
		return nil, err
	}

	return &binder{
		Logger:          logger,
		Config:          conf,
		Orderbook:       orderbook,
		RenExSettlement: settlement,
	}, nil
}

// GetOrderMatch checks if a match is found and returns the match object. It
// keeps doing it until an order match is found or the waitTill time.
func (binder *binder) GetOrderMatch(orderID [32]byte, waitTill int64) (swap.Match, error) {
	if err := binder.verifyOrder(orderID, waitTill); err != nil {
		return swap.Match{}, err
	}
	binder.LogInfo(orderID, "Waiting for the match to be found on RenEx")
	for {
		matchDetails, err := binder.GetMatchDetails(&bind.CallOpts{}, orderID)
		if err != nil {
			if time.Now().Unix() > waitTill {
				return swap.Match{}, fmt.Errorf("Timed out")
			}
			time.Sleep(10 * time.Second)
			continue
		}
		if !matchDetails.Settled {
			time.Sleep(10 * time.Second)
			continue
		}

		priorityToken, err := token.TokenCodeToToken(matchDetails.PriorityToken)
		if err != nil {
			return swap.Match{}, err
		}
		secondaryToken, err := token.TokenCodeToToken(matchDetails.SecondaryToken)
		if err != nil {
			return swap.Match{}, err
		}
		binder.LogInfo(orderID, fmt.Sprintf("matched with (%s%s%s)", pickColor(matchDetails.MatchedID), base64.StdEncoding.EncodeToString(matchDetails.MatchedID[:]), white))
		if matchDetails.OrderIsBuy {
			return swap.Match{
				PersonalOrderID: orderID,
				ForeignOrderID:  matchDetails.MatchedID,
				SendValue:       matchDetails.PriorityVolume.Add(matchDetails.PriorityVolume, matchDetails.PriorityFee).String(),
				ReceiveValue:    matchDetails.SecondaryVolume.Add(matchDetails.SecondaryVolume, matchDetails.SecondaryFee).String(),
				SendToken:       priorityToken,
				ReceiveToken:    secondaryToken,
			}, nil
		}
		return swap.Match{
			PersonalOrderID: orderID,
			ForeignOrderID:  matchDetails.MatchedID,
			SendValue:       matchDetails.SecondaryVolume.Add(matchDetails.SecondaryVolume, matchDetails.SecondaryFee).String(),
			ReceiveValue:    matchDetails.PriorityVolume.Add(matchDetails.PriorityVolume, matchDetails.PriorityFee).String(),
			SendToken:       secondaryToken,
			ReceiveToken:    priorityToken,
		}, nil
	}
}

func (binder *binder) verifyOrder(orderID [32]byte, waitTill int64) error {
	for {
		addr, err := binder.Orderbook.OrderTrader(&bind.CallOpts{}, orderID)
		if err != nil || addr.String() == "0x0000000000000000000000000000000000000000" {
			if time.Now().Unix() > waitTill {
				return fmt.Errorf("Timed out")
			}
			time.Sleep(10 * time.Second)
			continue
		}
		for _, authorizedAddr := range binder.AuthorizedAddresses {
			if strings.ToLower(addr.String()) == strings.ToLower(authorizedAddr) {
				return nil
			}
			binder.LogInfo(orderID, fmt.Sprintf("Expected submitting Trader Address %s to equal Authorized trader Address %s\n", addr.String(), authorizedAddr))
		}
	}
}

func pickColor(orderID [32]byte) string {
	return fmt.Sprintf("\033[3%dm", int64(orderID[0])%7)
}

const white = "\033[m"
