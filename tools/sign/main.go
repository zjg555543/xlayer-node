package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"net/http"
	"os"

	"github.com/0xPolygonHermez/zkevm-data-streamer/log"
	"github.com/0xPolygonHermez/zkevm-node/tools/sign/config"
	"github.com/0xPolygonHermez/zkevm-node/tools/sign/service"
)

const (
	appName  = "zkevm-data-streamer-tool" //nolint:gosec
	appUsage = "zkevm datastream tool"
)

var (
	configFileFlag = cli.StringFlag{
		Name:        config.FlagCfg,
		Aliases:     []string{"c"},
		Usage:       "Configuration `FILE`",
		DefaultText: "./config/tool.config.toml",
		Required:    true,
	}
)

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Usage = appUsage

	app.Commands = []*cli.Command{
		{
			Name:    "http",
			Aliases: []string{},
			Usage:   "Generate stream file from scratch",
			Action:  HttpService,
			Flags: []cli.Flag{
				&configFileFlag,
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func HttpService(cliCtx *cli.Context) error {
	c, err := config.Load(cliCtx)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	log.Init(c.Log)
	srv := service.NewServer(c, cliCtx.Context)
	http.HandleFunc("/priapi/v1/assetonchain/ecology/querySignDataByOrderNo", srv.GetSignDataByOrderNo)
	http.HandleFunc("/priapi/v1/assetonchain/ecology/ecologyOperate", srv.PostSignDataByOrderNo)

	log.Infof("%v,%v,%v,%v,%v,", c.L1.PolygonZkEVMAddress, c.L1.RPC, c.L1.ChainId, c.L1.SeqPrivateKey, c.L1.AggPrivateKey)
	log.Infof("%v", c.Port)

	// 启动 HTTP 服务器
	port := c.Port
	fmt.Printf("Server is running on :%d\n", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}

	return nil
}
