package logging

import (
	"github.com/sirupsen/logrus"
	cli "github.com/virtual-kubelet/node-cli"
	clilogrus "github.com/virtual-kubelet/node-cli/logrus"
	vklog "github.com/virtual-kubelet/virtual-kubelet/log"
	vklogrus "github.com/virtual-kubelet/virtual-kubelet/log/logrus"
)

func Options() []cli.Option {
	c := &clilogrus.Config{LogLevel: "info"}
	return []cli.Option{
		cli.WithPersistentFlags(c.FlagSet()),
		cli.WithPersistentPreRunCallback(logInitFunc(c)),
	}
}

func logInitFunc(c *clilogrus.Config) func() error {
	return func() error {
		logger := logrus.New()
		vklog.L = vklogrus.FromLogrus(logrus.NewEntry(logger))
		return clilogrus.Configure(c, logger)
	}
}
