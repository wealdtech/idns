package main

import (
	"flag"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	listener "github.com/wealdtech/go-eth-listener"
	"github.com/wealdtech/go-eth-listener/handlers"
)

// IDNSConfig is the configuration for IDNS
type IDNSConfig struct {
	Connection  *ethclient.Client
	OutputDir   string
	IPFSGateway string
}

func main() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	//logrus.SetLevel(logrus.InfoLevel)
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Info("Starting listener")

	// Options
	var connection string
	flag.StringVar(&connection, "connection", "", "Path to connection")
	var from string
	flag.StringVar(&from, "from", "", "Block from which to start (default current block)")
	var dir string
	flag.StringVar(&dir, "dir", "", "Output directory")
	var ipfsGateway string
	flag.StringVar(&ipfsGateway, "gateway", "", "IPFS gateway")
	flag.Parse()

	if connection == "" {
		fmt.Println("--connection is required")
		os.Exit(1)
	}

	if dir == "" {
		fmt.Println("--dir is required")
		os.Exit(1)
	}

	if ipfsGateway == "" {
		fmt.Println("--gateway is required")
		os.Exit(1)
	}

	client, err := ethclient.Dial(connection)
	if err != nil {
		panic(err)
	}

	config := &listener.Config{
		Connection:   client,
		Delay:        2,
		Timeout:      5 * time.Second,
		PollInterval: 1 * time.Minute,
		Extra: &IDNSConfig{
			Connection:  client,
			OutputDir:   dir,
			IPFSGateway: ipfsGateway,
		},
	}

	if from != "" {
		from, err := strconv.ParseInt(from, 10, 64)
		if err != nil {
			panic(err)
		}
		config.From = big.NewInt(from)
	}
	config.InitHandlers = IDNSInit(handlers.LogInit(nil))
	//config.BlkHandlers = handlers.LogBlk(nil)
	//config.TxHandlers = handlers.LogTx(nil)
	//config.EventHandlers = IDNSEvent(handlers.LogEvent(nil))
	config.EventHandlers = IDNSEvent(nil)

	listener.Listen(config)
}
