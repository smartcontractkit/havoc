package havoc_example

import (
	"github.com/rs/zerolog"
	"github.com/smartcontractkit/havoc"
	"github.com/stretchr/testify/require"
	"testing"
)

func createMonkey(t *testing.T, l zerolog.Logger, namespace string) *havoc.Controller {
	havoc.SetGlobalLogger(l)
	cfg, err := havoc.ReadConfig("config.toml")
	require.NoError(t, err)
	c, err := havoc.NewController(cfg)
	err = c.GenerateSpecs(namespace)
	require.NoError(t, err)
	return c
}

func TestMyLoad(t *testing.T) {
	/* my testing logger */
	l := havoc.L
	/* my load test preparation here */
	/* wrapping with chaos monkey */
	monkey := createMonkey(t, l, "my namespace, get it from config")
	go monkey.Run()
	/* my test runs and ends */
	errs := monkey.Stop()
	require.Len(t, errs, 0)
}
