package k8schaos

import (
	"context"
	"time"

	"github.com/smartcontractkit/chainlink-testing-framework/client"
)

type AnvilChaos struct {
	anvilClient *client.RPCClient
	chaosFun    func(ctx context.Context) error // Function to execute the actual anvil chaos
	listeners   []AnvilChaosListener

	BaseChaos
}

type AnvilChaosOpts struct {
	AnvilClient *client.RPCClient
	ChaosFun    func(ctx context.Context) error // Function to execute the actual anvil chaos
	Listeners   []AnvilChaosListener
	Description string
	DelayCreate time.Duration // Delay before creating the chaos object
}

func NewAnvilChaos(opts AnvilChaosOpts) (*AnvilChaos, error) {
	bc, err := NewBaseChaos(BaseChaosOpts{
		Description: opts.Description,
		DelayCreate: opts.DelayCreate,
	})

	return &AnvilChaos{BaseChaos: bc, anvilClient: opts.AnvilClient}, err
}

func (c *AnvilChaos) CreateNow() error {
	c.logger.Info().Msg("Creating Anvil chaos object")
	return c.chaosFun(context.Background())
}

func (c *AnvilChaos) OnChaosCreated() error {
	return nil
}

func (c AnvilChaos) OnChaosStarted() {
	for _, listener := range c.listeners {
		listener.OnChaosStarted(c)
	}
}

func (c *AnvilChaos) AddListener(listener AnvilChaosListener) {
	c.listeners = append(c.listeners, listener)
}
