package main

import (
	"context"
	"strings"

	cli "github.com/virtual-kubelet/node-cli"
	"github.com/virtual-kubelet/node-cli/opts"
	"github.com/virtual-kubelet/virtual-kubelet/log"

	"github.com/anqur/unitlet"
	"github.com/anqur/unitlet/internal/logging"
)

const k8sVersion = "v1.19.10"

var version, buildTime string

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = cli.ContextWithCancelOnSignal(ctx)

	o := opts.New()
	o.Provider = unitlet.ProviderName
	o.Version = strings.Join([]string{k8sVersion, unitlet.ProviderName, version}, "-")

	node, err := cli.New(
		ctx,
		append(
			[]cli.Option{
				cli.WithBaseOpts(o),
				cli.WithCLIVersion(version, buildTime),
				cli.WithProvider(unitlet.ProviderName, unitlet.New),
			},
			logging.Options()...,
		)...,
	)
	if err != nil {
		log.L.Fatal(err)
	}

	if err := node.Run(ctx); err != nil {
		log.L.Fatal(err)
	}
}
