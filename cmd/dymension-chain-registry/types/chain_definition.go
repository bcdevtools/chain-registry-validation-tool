package types

import (
	"fmt"
	"strings"
)

type ChainDefinition struct {
	ChainId       string                        `json:"chainId"`
	ChainName     string                        `json:"chainName"`
	RpcUrls       any                           `json:"rpc"`
	RestUrls      any                           `json:"rest"`
	Bech32Prefix  string                        `json:"bech32Prefix"`
	WebSite       string                        `json:"website,omitempty"`
	DA            string                        `json:"da"`
	EVM           *EvmChainDefinition           `json:"evm,omitempty"`
	Currencies    []CurrencyChainDefinition     `json:"currencies,omitempty"`
	CoinType      int64                         `json:"coinType"`
	GasAdjustment float64                       `json:"gasAdjustment,omitempty"`
	FaucetUrl     string                        `json:"faucetUrl,omitempty"`
	IBC           *IbcChainDefinition           `json:"ibc,omitempty"`
	GasPriceSteps *GasPriceStepsChainDefinition `json:"gasPriceSteps,omitempty"`
	Logo          string                        `json:"logo"`
	Type          string                        `json:"type"`
	Active        bool                          `json:"active,omitempty"`
	Analytics     bool                          `json:"analytics,omitempty"`
	CollectData   bool                          `json:"collectData,omitempty"`
	Goldberg      bool                          `json:"goldberg,omitempty"`
	AvailAddress  string                        `json:"availAddress,omitempty"`
}

type EvmChainDefinition struct {
	ChainId string `json:"chainId"`
	RpcUrls any    `json:"rpc"`
}

type CurrencyChainDefinition struct {
	DisplayDenom      string `json:"displayDenom"`
	BaseDenom         string `json:"baseDenom"`
	IbcRepresentation string `json:"ibcRepresentation"`
	BridgeDenom       string `json:"bridgeDenom"`
	Decimals          int64  `json:"decimals"`
	Logo              string `json:"logo,omitempty"`
	Type              string `json:"type"`
}

type IbcChainDefinition struct {
	Timeout    int64  `json:"timeout"`
	HubChannel string `json:"hubChannel"`
	Channel    string `json:"channel"`

	// https://github.com/dymensionxyz/chain-registry/blob/main/mainnet/ethereum/ethereum.json
	AllowedDenoms []string `json:"allowedDenoms"` // Must be existing in the currencies of the chain
}

type GasPriceStepsChainDefinition struct {
	Low     float64 `json:"low"`
	Average float64 `json:"average"`
	High    float64 `json:"high"`
}

func (cd ChainDefinition) GetRpcUrls() ([]string, error) {
	return getUrlsFromDynamicValue(cd.RpcUrls)
}

func (cd ChainDefinition) GetRestUrls() ([]string, error) {
	return getUrlsFromDynamicValue(cd.RestUrls)
}

func (cd ChainDefinition) GetEvmRpcUrls() ([]string, error) {
	if cd.EVM == nil {
		return nil, fmt.Errorf("EVM chain definition is not set")
	}
	return getUrlsFromDynamicValue(cd.EVM.RpcUrls)
}

func (cd ChainDefinition) IsRollAppChain() bool {
	return strings.EqualFold(cd.Type, "RollApp") || (cd.DA != "" && cd.Type == "")
}

func (cd ChainDefinition) IsEvmRollApp() bool {
	return cd.IsRollAppChain() && (cd.EVM != nil || cd.CoinType == 60)
}

func (cd ChainDefinition) IsDaAvail() bool {
	return strings.EqualFold(cd.DA, "Avail")
}

func getUrlsFromDynamicValue(dynamicValue any) (urls []string, err error) {
	if dynamicValue != nil {
		if strRpcUrl, ok := dynamicValue.(string); ok {
			urls = []string{strRpcUrl}
		} else if strRpcUrls, ok := dynamicValue.([]any); ok {
			for _, url := range strRpcUrls {
				if strUrl, ok := url.(string); ok {
					urls = append(urls, strUrl)
				} else {
					err = fmt.Errorf("url must be string, got %T", url)
					return
				}
			}
		} else {
			err = fmt.Errorf("url must be string or []string, got %T", dynamicValue)
			return
		}
	}

	return
}
