package main

import (
	"bytes"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/multiformats/go-multiaddr"
	"github.com/sirupsen/logrus"
	"github.com/wealdtech/go-ens/v3"
	"github.com/wealdtech/go-eth-listener/handlers"
	"github.com/wealdtech/go-eth-listener/shared"
)

var log = logrus.WithField("module", "idns")

// IDNSInit is called when the IDNS module is initialised.
func IDNSInit(h handlers.InitHandler) handlers.InitHandler {
	return handlers.InitHandlerFunc(func(actx *shared.AppContext) {
		log.Info("Initialised")
		if h != nil {
			h.Handle(actx)
		}
	})
}

// IDNSShutdown is called when the IDNS module is shut down.
func IDNSShutdown(h handlers.ShutdownHandler) handlers.ShutdownHandler {
	return handlers.ShutdownHandlerFunc(func(actx *shared.AppContext) {
		log.Info("Shutdown")
		if h != nil {
			h.Handle(actx)
		}
	})
}

var dnsZonehashChangedTopic = []byte{0x8f, 0x15, 0xed, 0x4b, 0x72, 0x3e, 0xf4, 0x28, 0xf2, 0x50, 0x96, 0x1d, 0xa8, 0x31, 0x56, 0x75, 0xb5, 0x07, 0x04, 0x67, 0x37, 0xe1, 0x93, 0x19, 0xfc, 0x1a, 0x4d, 0x81, 0xbf, 0xe8, 0x7f, 0x85}

// IDNSEvent is called when the IDNS module receives an event.
func IDNSEvent(nextHandler handlers.EventHandler) handlers.EventHandler {
	return handlers.EventHandlerFunc(func(actx *shared.AppContext, event *types.Log) {
		if nextHandler != nil {
			defer nextHandler.Handle(actx, event)
		}

		log = log.WithField("event", event)

		if len(event.Topics) != 2 ||
			!bytes.Equal(dnsZonehashChangedTopic, event.Topics[0].Bytes()) {
			// Not our event
			return
		}

		// Need to obtain the last and current zone hashes
		lastZoneHashLength := new(big.Int).SetBytes(event.Data[64:96]).Int64()
		lastZoneHash := event.Data[96 : 96+lastZoneHashLength]
		zoneHashStart := 96 + 32*((lastZoneHashLength+31)/32)
		zoneHashLength := new(big.Int).SetBytes(event.Data[zoneHashStart : zoneHashStart+32]).Int64()
		zoneHash := event.Data[zoneHashStart+32 : zoneHashStart+32+zoneHashLength]

		var nameHash [32]byte
		copy(nameHash[:], event.Topics[1].Bytes())

		if bytes.Equal(zoneHash, []byte{}) {
			contentHash, err := ens.ContenthashToString(lastZoneHash)
			if err != nil {
				log.WithError(err).WithField("hash", fmt.Sprintf("%x", lastZoneHash)).Warn("failed to parse removal content hash")
				return
			}

			addr, err := multiaddr.NewMultiaddr(contentHash)
			if err != nil {
				log.WithError(err).Warn("failed to parse removal multiaddr")
				return
			}
			// Kick off the clear job.
			go clear(&event.Address, addr, nameHash, actx.Extra.(*IDNSConfig))
		} else {
			contentHash, err := ens.ContenthashToString(zoneHash)
			if err != nil {
				log.WithError(err).WithField("hash", fmt.Sprintf("%x", zoneHash)).Warn("failed to parse update content hash")
				return
			}

			addr, err := multiaddr.NewMultiaddr(contentHash)
			if err != nil {
				log.WithError(err).Warn("failed to parse update multiaddr")
				return
			}
			// Kick off the fetch job.
			go fetch(&event.Address, addr, nameHash, actx.Extra.(*IDNSConfig))
		}
	})
}
