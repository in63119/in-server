package post

import "encoding/json"

type NftAttribute struct {
	TraitType   string          `json:"trait_type"`
	Value       json.RawMessage `json:"value"`
	DisplayType string          `json:"display_type,omitempty"`
}

type NftMetadata struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Image       string         `json:"image,omitempty"`
	ExternalURL string         `json:"external_url"`
	Attributes  []NftAttribute `json:"attributes"`
}
