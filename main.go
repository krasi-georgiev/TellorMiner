package main

import (
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/event"
	"github.com/pkg/errors"
	"github.com/tellor-io/telliot/pkg/contracts/tellor"
)

type ProfitChecker struct {
	client *ethclient.Client
}

func main() {
	client, err := ethclient.Dial("NODE URL")
	if err != nil {
		log.Fatal(err)
	}

	self := &ProfitChecker{
		client: client,
	}

	self.monitorReward()
}

func (self *ProfitChecker) monitorReward() {
	var err error
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var sub event.Subscription
	events := make(chan *tellor.ITellorNonceSubmitted)

	for {
		sub, err = self.Sub(events)
		if err != nil {
			log.Print("msg", "initial subscribing to  events failed")
			<-ticker.C
			continue
		}
		break
	}

	for {
		select {
		case err := <-sub.Err():
			if err != nil {
				log.Print(
					"msg",
					" subscription error",
					"err", err)
			}

			// Trying to resubscribe until it succeeds.
			for {
				sub, err = self.Sub(events)
				if err != nil {
					log.Print("msg", "re-subscribing to  events failed")
					<-ticker.C
					continue
				}
				break
			}
			log.Print("msg", "re-subscribed to  events")
		case event := <-events:
			fmt.Println("vLog tranfer", event)
		}
	}
}

func (self *ProfitChecker) Sub(output chan *tellor.ITellorNonceSubmitted) (event.Subscription, error) {
	var tellorFilterer *tellor.ITellorFilterer
	tellorFilterer, err := tellor.NewITellorFilterer(common.HexToAddress("0x88dF592F8eb5D7Bd38bFeF7dEb0fBc02cf3778a0"), self.client)
	if err != nil {
		return nil, errors.Wrap(err, "getting ITellorFilterer instance")
	}
	sub, err := tellorFilterer.WatchNonceSubmitted(
		nil,
		output,
		nil,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "getting  channel")
	}
	return sub, nil
}
