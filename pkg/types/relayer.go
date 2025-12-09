package types

type RelayerStatus string

const (
	RelayerStatusReady      RelayerStatus = "Ready"
	RelayerStatusProcessing RelayerStatus = "Processing"
	RelayerStatusShutdown   RelayerStatus = "Shutdown"
)

type FirebaseRelayer struct {
	Address string        `json:"address"`
	Status  RelayerStatus `json:"status"`
}
