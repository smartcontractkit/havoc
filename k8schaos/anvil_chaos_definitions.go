package k8schaos

import (
	"context"
	"time"

	"github.com/smartcontractkit/chainlink-testing-framework/client"
)

func SetHeadAnvilChaos(anvilClient *client.RPCClient, blocksBack int, listeners []AnvilChaosListener) (*AnvilChaos, error) {
	return NewAnvilChaos(AnvilChaosOpts{
		AnvilClient: anvilClient,
		ChaosFun: func(ctx context.Context) error {
			return anvilClient.GethSetHead(blocksBack)
		},
		Description: "Set head of Anvil client",
		DelayCreate: 10 * time.Second,
		Listeners:   listeners,
	})
}

func CCIP_TEST() {
	srcClient := client.NewRPCClient("http://localhost:8545")

	chaos, err := SetHeadAnvilChaos(srcClient, 10, []AnvilChaosListener{})
	if err != nil {
		panic(err)
	}
	err = chaos.CreateSync(context.Background())
	if err != nil {
		panic(err)
	}
}
