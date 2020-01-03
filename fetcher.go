package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/miekg/dns"
	"github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
	"github.com/wealdtech/go-ens/v3"
)

// Clear clears a DNS zonefile.
func clear(contractAddress *common.Address, lastAddr multiaddr.Multiaddr, nodeHash [32]byte, config *IDNSConfig) {
	zonefile, err := fetchZoneFile(lastAddr, config)
	if err != nil {
		log.WithError(err).Warn("Failed to fetch zone file")
		return
	}
	zone, err := parseZone(zonefile, nodeHash)
	if err != nil {
		log.WithError(err).Warn("Failed to parse zone file")
		return
	}
	validOrigin, err := assertEventOrigin(config.Connection, zone, contractAddress)
	if err != nil {
		log.WithError(err).Warn("Failed to assert event origin")
		return
	}
	if !validOrigin {
		log.Warn("Invalid event origin")
	}

	// Delete the zone
	outPath := filepath.Join(config.OutputDir, fmt.Sprintf("db.%s", zone))
	log = log.WithField("path", outPath)
	err = os.Remove(outPath)
	if err != nil {
		log.WithError(err).WithField("path", outPath).Warn("Failed to delete zonefile")
	}
}

// fetch fetches a DNS zonefile from a remote address and stores it locally.
func fetch(contractAddress *common.Address, addr multiaddr.Multiaddr, nodeHash [32]byte, config *IDNSConfig) {
	zonefile, err := fetchZoneFile(addr, config)
	if err != nil {
		log.WithError(err).Warn("Failed to fetch zone file")
		return
	}

	zone, err := parseZone(zonefile, nodeHash)
	if err != nil {
		log.WithError(err).Warn("Failed to parse zone file")
		return
	}
	validOrigin, err := assertEventOrigin(config.Connection, zone, contractAddress)
	if err != nil {
		log.WithError(err).Warn("Failed to assert event origin")
		return
	}
	if !validOrigin {
		log.Warn("Invalid event origin")
	}

	// Write out the zonefile
	outPath := filepath.Join(config.OutputDir, fmt.Sprintf("db.%s", zone))
	log = log.WithField("path", outPath)
	file, err := os.Create(outPath)
	if err != nil {
		log.WithError(err).Warn("Failed to create zone file")
		return
	}
	defer file.Close()
	_, err = io.WriteString(file, zonefile)
	if err != nil {
		log.WithError(err).Warn("Failed to write zone file")
		return
	}
}

// fetchZoneFile fetches a zonefile.
func fetchZoneFile(addr multiaddr.Multiaddr, config *IDNSConfig) (string, error) {
	protocols := addr.Protocols()
	log = log.WithField("multiaddr", addr)
	if len(protocols) == 0 {
		return "", errors.New("failed to obtain multiaddr protocol")
	}
	// We are interested in the final protocol
	var zonefile string
	var err error
	switch protocols[len(protocols)-1].Code {
	case multiaddr.P_P2P:
		// IPFS
		zonefile, err = fetchFromIPFS(addr, config)
	default:
		log.WithField("protocol", protocols[len(protocols)-1].Name).Warn("Unhandled multiaddr protocol")
		return "", errors.New("Unknown multiaddr protocol")
	}

	if err != nil {
		return "", errors.Wrap(err, "Failed to obtain zonefile")
	}
	return zonefile, nil
}

// parseZone parses a zonefile and validates it.
func parseZone(zonefile string, nodeHash [32]byte) (string, error) {
	// Parse the zonefile
	zoneParser := dns.NewZoneParser(strings.NewReader(zonefile), "", "")
	if zoneParser == nil {
		return "", errors.New("Failed to create zone parser")
	}

	// We need to ensure the zonefile is for the appropriate zone; find the SOA record(s) to confirm
	soaMatch := false
	zone := ""
	soas := 0
	for rr, ok := zoneParser.Next(); ok; rr, ok = zoneParser.Next() {
		switch rr.(type) {
		case *dns.SOA:
			soas++
			soa := rr.(*dns.SOA)
			zone = strings.TrimSuffix(soa.Header().Name, ".")
			log = log.WithField("zone", zone)
			soaHash, err := ens.NameHash(zone)
			if err != nil {
				return "", errors.Wrap(err, "failed to create namehash for SOA")
			}
			soaMatch = bytes.Equal(soaHash[:], nodeHash[:])
		}
	}

	if soas == 0 {
		return "", errors.New("no SOA")
	}
	if soas > 1 {
		return "", errors.New("multiple SOAs")
	}
	if !soaMatch {
		return "", errors.New("mismatched SOAs")
	}

	return zone, nil
}

// assertEventOrigin ensures the event originated from the resolver for the domain.
func assertEventOrigin(conn *ethclient.Client, zone string, contractAddress *common.Address) (bool, error) {
	registry, err := ens.NewRegistry(conn)
	if err != nil {
		return false, errors.Wrap(err, "failed to obtain registry")
	}
	resolverAddress, err := registry.ResolverAddress(zone)
	if err != nil || resolverAddress == ens.UnknownAddress {
		return false, errors.Wrap(err, "failed to obtain resolver")
	}
	return bytes.Equal(resolverAddress.Bytes(), contractAddress.Bytes()), nil
}

func fetchFromIPFS(addr multiaddr.Multiaddr, config *IDNSConfig) (string, error) {
	path := fmt.Sprintf("%s%s", config.IPFSGateway, addr.String())
	// Strip double "//" if present (except after the protocol specifier)
	re := regexp.MustCompile(`([^:])//`)
	path = re.ReplaceAllString(path, "$1/")

	// Swap out /p2p/ for /ipfs/
	re2 := regexp.MustCompile(`/p2p/`)
	path = re2.ReplaceAllString(path, "/ipfs/")

	resp, err := http.Get(path)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
