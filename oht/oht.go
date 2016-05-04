package oht

import (
	"os"
	"os/signal"
	"syscall"

	"./../accounts"
	"./../contacts"
	"./common"
	"./network"
)

type OHT struct {
	// Should client name,version and other information be cached here?
	Interface *Interface
	config    *Config
	tor       *network.TorProcess
	// Channel for shutting down the oht
	shutdownChan chan bool
	// This should be in its own sub-module
	accountManager *accounts.Manager
	//protocolManager *ProtocolManager -- will this be useful?
	//eventMux *event.TypeMux
}

func NewOHT(torListenPort, torSocksPort, torControlPort, torWebUIPort int) *OHT {
	// Initialize Data Directory
	if !common.FileExist(common.DefaultDataDir()) {
		os.MkdirAll(common.DefaultDataDir(), os.ModePerm)
	}
	// Check if config exists,
	//  if true  - initialize it
	//  if false - load it
	// Set defaults for torPorts to use if not specified
	initializeConfig()
	config := &Config{
		clientName:         "oth",
		clientMajorVersion: 0,
		clientMinorVersion: 1,
		clientPatchVersion: 0,
	}

	// Should starting tor be a separate function? Functions to control Tor will be required...
	contacts.InitializeContacts()
	tor := network.InitializeTor(torListenPort, torSocksPort, torControlPort, torWebUIPort)
	accountsManager := &accounts.Manager{} // Placeholder

	// Start P2P Networking
	go network.Manager.Start(tor.ListenPort)

	// Define a clean shutdown process
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		tor.Shutdown()
		os.Exit(1)
	}()

	return &OHT{
		Interface:      NewInterface(config, tor, accountsManager),
		config:         config,
		tor:            tor,
		accountManager: accountsManager,
	}
}