package system

import (
	"cinco-paus/seismic/client"
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"pkg.world.dev/world-engine/cardinal"
	"pkg.world.dev/world-engine/sign"
)

var (
	proofRequestCh  = make(chan client.ProofRequest)
	proofReturnCh   = make(chan client.ProofReqResponse)
	revealRequestCh = make(chan int)
	revealReturnCh  = make(chan int)
)

func Initialize(world *cardinal.World) *client.SeismicClient {
	fulFillMsg, ok := world.GetMessageByFullName("game.fulfill-create-game")
	if !ok {
		fmt.Printf("error: no 'fulfill-create-game' message")
		return nil
	}
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			select {
			case proofRes := <-proofReturnCh:
				if proofRes.Success {
					ok := client.Verify(proofRes.Proof)
					fmt.Println("verification res:", ok)
				}

				// TODO: fix this signature
				sig, err := sign.NewTransaction(privateKey, "Seismic.Systems", world.Namespace(), 0, `{}`)
				if err != nil {
					fmt.Printf("failed to sign new tx: %v", err)
				}
				world.AddTransaction(fulFillMsg.ID(), `{}`, sig)

			case revealRes := <-revealReturnCh:
				fmt.Println(revealRes)
			}
		}
	}()

	return client.New(proofRequestCh, proofReturnCh, revealRequestCh, revealReturnCh)
}