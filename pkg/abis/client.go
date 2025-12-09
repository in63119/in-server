package abis

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"

	"in-server/pkg/types"
)

type SupportedEnv string

const (
	EnvDevelopment SupportedEnv = "development"
	EnvProduction  SupportedEnv = "production"
)

type ContractArtifact struct {
	Address string          `json:"address"`
	ABI     json.RawMessage `json:"abi"`
}

//go:embed kaia/test/development/*.json
var devFiles embed.FS

//go:embed kaia/test/production/*.json
var prodFiles embed.FS

var contractFileMap = map[types.ContractName]string{
	types.AUTHSTORAGE:       "AuthStorage.json",
	types.POSTSTORAGE:       "PostStorage.json",
	types.INFORWARDER:       "InForwarder.json",
	types.RELAYERMANAGER:    "RelayerManager.json",
	types.VISITORSTORAGE:    "VisitorStorage.json",
	types.YOUTUBESTORAGE:    "YoutubeStorage.json",
	types.SUBSCRIBERSTORAGE: "SubscriberStorage.json",
}

func Get(env string) (map[types.ContractName]ContractArtifact, error) {
	base := "kaia/test/development/"
	switch strings.ToLower(strings.TrimSpace(env)) {
	case string(EnvProduction):
		base = "kaia/test/production/"
		return loadArtifacts(prodFiles, base)
	default:
		return loadArtifacts(devFiles, base)
	}
}

func loadArtifacts(fs embed.FS, base string) (map[types.ContractName]ContractArtifact, error) {
	out := make(map[types.ContractName]ContractArtifact, len(contractFileMap))
	for name, file := range contractFileMap {
		raw, err := fs.ReadFile(base + file)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", file, err)
		}

		var art ContractArtifact
		if err := json.Unmarshal(raw, &art); err != nil {
			return nil, fmt.Errorf("parse %s: %w", file, err)
		}

		out[name] = art
	}
	return out, nil
}
