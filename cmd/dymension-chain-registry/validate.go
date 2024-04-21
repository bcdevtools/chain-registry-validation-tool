package dymension_chain_registry

import (
	"encoding/json"
	"fmt"
	valtypes "github.com/bcdevtools/chain-registry-validation-tool/cmd/dymension-chain-registry/types"
	"github.com/bcdevtools/chain-registry-validation-tool/utils"
	"github.com/spf13/cobra"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	flagMainnet                   = "mainnet"
	flagTestnet                   = "testnet"
	flagDevnet                    = "devnet"
	flagInternalDevnet            = "internal-devnet"
	flagStopOnFirstErr            = "stop-on-error"
	flagAdditionChainTypesAllowed = "addition-chain-types-allowed"
)

var validationErrors []string

func GetValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "validate [repo-dir]",
		Aliases: []string{"v"},
		Short:   "Validate Dymension chain-registry",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var validateMainnet, validateTestnet, validateDevnet, validateInternalDevnet bool

			if cmd.Flags().Changed(flagMainnet) {
				validateMainnet = true
			}
			if cmd.Flags().Changed(flagTestnet) {
				validateTestnet = true
			}
			if cmd.Flags().Changed(flagDevnet) {
				validateDevnet = true
			}
			if cmd.Flags().Changed(flagInternalDevnet) {
				validateInternalDevnet = true
			}

			stopOnFirstError := cmd.Flags().Changed(flagStopOnFirstErr)

			additionalChainTypesAllowed, _ := cmd.Flags().GetStringArray(flagAdditionChainTypesAllowed)

			if !validateMainnet && !validateTestnet && !validateDevnet && !validateInternalDevnet {
				// no flag provided, validate all
				validateMainnet = true
				validateTestnet = true
				validateDevnet = true
				validateInternalDevnet = true
			}

			fmt.Printf("Going to validate")
			if validateMainnet {
				fmt.Printf(" %s", valtypes.ValidateMainnet)
			}
			if validateTestnet {
				fmt.Printf(" %s", valtypes.ValidateTestnet)
			}
			if validateDevnet {
				fmt.Printf(" %s", valtypes.ValidateDevnet)
			}
			if validateInternalDevnet {
				fmt.Printf(" %s", valtypes.ValidateInternalDevnet)
			}
			fmt.Println()

			repoDir := args[0]

			di, err := os.Stat(repoDir)
			if err != nil {
				if os.IsNotExist(err) {
					utils.PrintlnStdErr("ERR: Provided 'chain-registry' repository path does not exists")
					os.Exit(1)
				}
				utils.PrintlnStdErr("ERR: Failed to get stat of provided 'chain-registry' repository path:", err)
				os.Exit(1)
			}

			if !di.IsDir() {
				utils.PrintlnStdErr("ERR: Provided 'chain-registry' repository path is not a directory")
				os.Exit(1)
			}

			if validateMainnet {
				validateChainRegistry(repoDir, valtypes.ValidateMainnet, stopOnFirstError, additionalChainTypesAllowed)
			}
			if validateTestnet {
				validateChainRegistry(repoDir, valtypes.ValidateTestnet, stopOnFirstError, additionalChainTypesAllowed)
			}
			if validateDevnet {
				validateChainRegistry(repoDir, valtypes.ValidateDevnet, stopOnFirstError, additionalChainTypesAllowed)
			}
			if validateInternalDevnet {
				validateChainRegistry(repoDir, valtypes.ValidateInternalDevnet, stopOnFirstError, additionalChainTypesAllowed)
			}

			if len(validationErrors) > 0 {
				utils.PrintlnStdErr("Errors:")
				for _, err := range validationErrors {
					utils.PrintlnStdErr(">", err)
				}
				utils.PrintlnStdErr("Total", len(validationErrors), "issues found!")
				os.Exit(1)
			}

			fmt.Println("Passed!")
		},
	}

	cmd.Flags().Bool(flagMainnet, false, "validate mainnet records only")
	cmd.Flags().Bool(flagTestnet, false, "validate testnet records only")
	cmd.Flags().Bool(flagDevnet, false, "validate devnet records only")
	cmd.Flags().Bool(flagInternalDevnet, false, "validate internal-devnet records only")
	cmd.Flags().BoolP(flagStopOnFirstErr, "e", false, "stop on first error")
	cmd.Flags().StringArray(flagAdditionChainTypesAllowed, nil, "allow additional chain types")

	return cmd
}

func validateChainRegistry(repoDir string, target valtypes.ValidateTarget, stopOnFirstErr bool, additionalChainTypesAllowed []string) {
	fmt.Println("Validating group", target.String(), "...")

	var workingChain string
	var workingFile string
	markErr := func(a ...any) {
		prefixes := []any{"ERR:", fmt.Sprintf("[group:%s]", target.String())}
		if len(workingChain) > 0 {
			prefixes = append(prefixes, fmt.Sprintf("[chain:%s]", workingChain))
		}
		prefixes = append(prefixes, "Validation failed!")

		errMsgParts := append(prefixes, a...)

		utils.PrintlnStdErr(errMsgParts...)

		if len(workingFile) > 0 {
			utils.PrintlnStdErr("File:", workingFile)
			utils.PrintlnStdErr()
			errMsgParts = append(errMsgParts, ", File: ", workingFile)
		} else {
			utils.PrintlnStdErr()
		}

		if stopOnFirstErr {
			os.Exit(1)
		}

		validationErrors = append(validationErrors, fmt.Sprint(errMsgParts...))
	}

	subDirPath := path.Join(repoDir, target.SubDirectoryName())
	di, err := os.Stat(subDirPath)
	if err != nil {
		if os.IsNotExist(err) {
			markErr("Missing required directory", target.SubDirectoryName(), "at", subDirPath)
			os.Exit(1)
		}
		markErr("Failed to get stat of", subDirPath, "directory:", err)
		os.Exit(1)
	}
	if !di.IsDir() {
		markErr("Expected target path is not a directory:", subDirPath)
		os.Exit(1)
	}

	uniqueChainIdTracker := make(map[string]string)

	_ = filepath.WalkDir(subDirPath, func(filePath string, d os.DirEntry, _ error) error {
		if !d.IsDir() {
			return nil
		}
		if strings.HasSuffix(filePath, fmt.Sprintf("%c%s", os.PathSeparator, target.SubDirectoryName())) {
			return nil
		}

		spl := strings.Split(filePath, fmt.Sprintf("%c%s%c", os.PathSeparator, target.SubDirectoryName(), os.PathSeparator))
		if len(spl) < 2 {
			return nil
		}
		spl = strings.Split(spl[1], string(os.PathSeparator))
		if len(spl) > 1 {
			// skip sub-dir of chains
			return nil
		}

		//fmt.Println("> Validating", filePath)

		workingChain = spl[0]

		chainDefinitionFile := path.Join(filePath, workingChain+".json")

		_, err := os.Stat(chainDefinitionFile)
		if err != nil {
			if os.IsNotExist(err) {
				markErr("Missing required file", chainDefinitionFile)
				return nil
			}
			markErr("Failed to get stat of", chainDefinitionFile, "file:", err)
			return nil
		}

		workingFile = chainDefinitionFile

		bzChainDefinition, err := os.ReadFile(chainDefinitionFile)
		if err != nil {
			markErr("Failed to read chain definition file:", err)
			return nil
		}

		var cd valtypes.ChainDefinition
		err = json.Unmarshal(bzChainDefinition, &cd)
		if err != nil {
			markErr("Failed to unmarshal chain definition file:", err)
			return nil
		}

		if existing, found := uniqueChainIdTracker[cd.ChainId]; found {
			markErr("Duplicated chain id found:", cd.ChainId, "in", existing, "and", workingChain)
			return nil
		}
		uniqueChainIdTracker[cd.ChainId] = workingChain

		if !isValidChainId(cd.ChainId, cd.IsRollAppChain() && cd.EVM != nil) {
			markErr("Bad chain id:", cd.ChainId)
		}

		if !isValidChainName(cd.ChainName) {
			markErr("Bad chain name:", cd.ChainName)
		}

		rpcUrls, err := cd.GetRpcUrls()
		if err != nil {
			markErr("Failed to get RPC urls:", err)
		} else if !isValidUrls(rpcUrls) {
			markErr("Bad RPC urls:", rpcUrls)
		}

		restUrls, err := cd.GetRestUrls()
		if err != nil {
			markErr("Failed to get REST urls:", err)
		} else if !isValidUrls(restUrls) {
			markErr("Bad REST urls:", restUrls)
		}

		beRpcUrls, err := cd.GetBeRpcUrls()
		if err != nil {
			markErr("Failed to get Be RPC urls:", err)
		} else if !isValidUrls(beRpcUrls) {
			markErr("Bad Be RPC urls:", beRpcUrls)
		}

		if cd.IsRollAppChain() {
			if cd.Bech32Prefix == "" {
				markErr("Bech32 prefix is required for RollApp chains")
			}
		}
		if cd.Bech32Prefix != "" {
			if !isValidBech32Prefix(cd.Bech32Prefix) {
				markErr("Bad Bech32 prefix:", cd.Bech32Prefix)
			}
		}

		if !isValidOptionalWebsiteUrl(cd.WebSite) {
			markErr("Bad website url:", cd.WebSite)
		}

		if !isValidDA(cd) {
			markErr("Bad DA:", cd.DA)
		}

		if cd.CoinType == 60 && cd.IsRollAppChain() && cd.EVM == nil {
			markErr("\"evm\" is required for RollApp EVM chains")
		}

		if cd.EVM != nil {
			evmRpcUrls, err := cd.GetEvmRpcUrls()
			if err != nil {
				markErr("Failed to get EVM RPC urls:", err)
			} else if !isValidUrls(evmRpcUrls) {
				markErr("Bad EVM RPC urls:", evmRpcUrls)
			}

			if !isValidEvmHexChainId(cd) {
				markErr("Bad EVM hex chain id:", cd.EVM.ChainId)
			}
		}

		if len(cd.Currencies) > 0 {
			if valid, identity := isValidCurrencies(cd.Currencies, filePath); !valid {
				if identity == "" {
					markErr("Bad currencies")
				} else {
					markErr("Bad currencies:", identity)
				}
			}
		} else {
			markErr("Currencies is required")
		}

		if cd.IsEvmRollApp() {
			if cd.CoinType != 60 {
				markErr("Coin type must be 60 for EVM RollApp chains")
			}
		} else if !isValidCoinType(cd.CoinType) {
			markErr("Bad coin type:", cd.CoinType)
		}

		if !isValidGasAdjustment(cd.GasAdjustment) {
			markErr("Bad gas adjustment:", cd.GasAdjustment)
		}

		if !isValidOptionalWebsiteUrl(cd.FaucetUrl) {
			markErr("Bad faucet url:", cd.FaucetUrl)
		}

		if cd.IBC != nil {
			if !isValidIbc(cd.IBC) {
				markErr("Bad IBC:", cd.IBC)
			}
		}

		if cd.GasPriceSteps != nil {
			if !isValidGasPriceSteps(cd.GasPriceSteps) {
				markErr("Bad gas price steps:", cd.GasPriceSteps)
			}
		}

		if !isValidLogo(cd.Logo, filePath) {
			markErr("Bad chain logo:", cd.Logo)
		}

		if !isValidChainType(cd.Type, additionalChainTypesAllowed) {
			markErr("Bad chain type:", cd.Type)
		}

		if cd.Goldberg && cd.DA != "Avail" {
			markErr("Goldberg when set, DA must be Avail")
		}

		if !isValidAvailAddress(cd.AvailAddress, cd.DA) {
			markErr("Bad avail address:", cd.AvailAddress)
		}

		return nil
	})
}

func isValidAvailAddress(availAddress string, da string) bool {
	if da != "Avail" {
		if availAddress != "" {
			utils.PrintlnStdErr("ERR: Avail address is only available if DA is Avail")
			return false
		}
		return true
	}

	if availAddress == "" {
		return true
	}

	if strings.Contains(availAddress, " ") {
		utils.PrintlnStdErr("ERR: Avail address must not contains space")
		return false
	}

	if !strings.HasPrefix(availAddress, "5") {
		utils.PrintlnStdErr("ERR: Avail address must start with 5")
		return false
	}

	if !regexp.MustCompile(`^5[a-zA-Z\d]+$`).MatchString(availAddress) {
		utils.PrintlnStdErr("ERR: Avail address must starts with 5, followed by alphanumeric characters")
		return false
	}

	if len(availAddress) != 48 {
		utils.PrintlnStdErr("ERR: Avail address must be 48 characters long")
		return false
	}

	return true
}

func isValidChainType(chainType string, additionalChainTypesAllowed []string) bool {
	if chainType == "" {
		utils.PrintlnStdErr("ERR: Chain type is required")
		return false
	}

	switch chainType {
	case "RollApp", "Regular", "EVM", "Hub", "Solana":
		return true
	default:
		for _, ct := range additionalChainTypesAllowed {
			if ct == ct {
				return true
			}
		}
		utils.PrintlnStdErr("ERR: Not recognized chain type:", chainType, fmt.Sprintf("(consider provide into --%s flag)", flagAdditionChainTypesAllowed))
		return false
	}
}

func isValidLogo(logo string, chainPath string) bool {
	if logo == "" {
		return true
	}
	logoPath := path.Join(chainPath, logo)
	_, err := os.Stat(logoPath)
	if err != nil {
		if os.IsNotExist(err) {
			utils.PrintlnStdErr("ERR: Logo file not found:", logoPath)
			return false
		}
		utils.PrintlnStdErr("ERR: Failed to get stat of logo file:", logoPath, err)
		return false
	}
	ext := strings.ToLower(filepath.Ext(logoPath))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".svg":
		return true
	default:
		utils.PrintlnStdErr("ERR: Logo file must be PNG, JPG, JPEG, or SVG:", logoPath)
		return false
	}
}

func isValidGasPriceSteps(gasPriceSteps *valtypes.GasPriceStepsChainDefinition) bool {
	if gasPriceSteps.Low <= 0 {
		utils.PrintlnStdErr("ERR: Gas price steps low must be positive")
		return false
	}
	if gasPriceSteps.Average <= 0 {
		utils.PrintlnStdErr("ERR: Gas price steps average must be positive")
		return false
	}
	if gasPriceSteps.High <= 0 {
		utils.PrintlnStdErr("ERR: Gas price steps high must be positive")
		return false
	}
	if gasPriceSteps.Low > gasPriceSteps.Average {
		utils.PrintlnStdErr("ERR: Gas price steps low must not exceed average")
		return false
	}
	if gasPriceSteps.Average > gasPriceSteps.High {
		utils.PrintlnStdErr("ERR: Gas price steps average must not exceed high")
		return false
	}
	return true
}

func isValidIbc(ibc *valtypes.IbcChainDefinition) bool {
	if ibc.Channel != "" {
		if ibc.Channel == "-" {
			// special case
		} else if !regexp.MustCompile(`^channel-\d+$`).MatchString(ibc.Channel) {
			utils.PrintlnStdErr("ERR: IBC channel must match format channel-<number>")
			return false
		}
	}
	if ibc.HubChannel != "" {
		if !regexp.MustCompile(`^channel-\d+$`).MatchString(ibc.HubChannel) {
			utils.PrintlnStdErr("ERR: IBC hub channel must match format channel-<number>")
			return false
		}
	}
	if ibc.HubChannel != "" && ibc.Channel == "" {
		utils.PrintlnStdErr("ERR: IBC channel is required if hub channel is set")
		return false
	}
	if ibc.Timeout < 0 {
		utils.PrintlnStdErr("ERR: IBC timeout must not be negative")
		return false
	}
	if len(ibc.AllowedDenoms) > 0 {
		uniquenessTracker := make(map[string]bool)
		for _, denom := range ibc.AllowedDenoms {
			if denom == "" {
				utils.PrintlnStdErr("ERR: IBC allowed denom must not be empty")
				return false
			}
			if strings.TrimSpace(denom) != denom {
				utils.PrintlnStdErr("ERR: IBC allowed denom must not have leading or trailing spaces")
				return false
			}
			if strings.Contains(denom, " ") {
				utils.PrintlnStdErr("ERR: IBC allowed denom must not contains space")
				return false
			}
			if strings.Contains(denom, "//") {
				utils.PrintlnStdErr("ERR: IBC allowed denom must not contains consecutive slashes")
				return false
			}
			if strings.Contains(denom, "--") {
				utils.PrintlnStdErr("ERR: IBC allowed denom must not contains consecutive dashes")
				return false
			}
			if strings.Contains(denom, "__") {
				utils.PrintlnStdErr("ERR: IBC allowed denom must not contains consecutive underscores")
				return false
			}
			if !regexp.MustCompile(`^[a-zA-Z\d-_/]+$`).MatchString(denom) {
				utils.PrintlnStdErr("ERR: IBC allowed denom must be alphanumeric, dash, underscore, or slash")
				return false
			}
			if _, found := uniquenessTracker[denom]; found {
				utils.PrintlnStdErr("ERR: Duplicated IBC allowed denom found:", denom)
				return false
			}
			uniquenessTracker[denom] = true
		}
	}
	return true
}

func isValidGasAdjustment(gasAdjustment float64) bool {
	if gasAdjustment == 0.0 {
		return true
	}
	if gasAdjustment < 0.0 {
		utils.PrintlnStdErr("ERR: Gas adjustment must be non-negative", gasAdjustment)
		return false
	}
	if gasAdjustment < 1.0 {
		utils.PrintlnStdErr("ERR: Gas adjustment must be at least 1.0", gasAdjustment)
		return false
	}
	return true
}

func isValidCoinType(coinType int64) bool {
	if coinType < 0 {
		utils.PrintlnStdErr("ERR: Coin type must be non-negative")
		return false
	}
	if coinType > 255 {
		utils.PrintlnStdErr("ERR: Coin type must not exceed 255")
		return false
	}
	return true
}

func isValidCurrencies(currencies []valtypes.CurrencyChainDefinition, chainPath string) (valid bool, identity string) {
	var foundMain bool

	uniqueBaseDenomTracker := make(map[string]bool)
	uniqueDisplayDenomTracker := make(map[string]bool)
	uniqueIbcRepresentationTracker := make(map[string]bool)

	for _, currency := range currencies {
		if !isValidCurrency(currency, chainPath) {
			var descCurrency string
			bz, err := json.Marshal(currency)
			if err != nil {
				descCurrency = fmt.Sprintln(currency)
			} else {
				descCurrency = string(bz)
			}

			utils.PrintlnStdErr("Bad currency:", descCurrency)

			return false, descCurrency
		}

		if currency.Type == "main" {
			if foundMain {
				utils.PrintlnStdErr("ERR: Duplicated main currency found")
				return false, currency.BaseDenom
			} else {
				foundMain = true
			}
		}

		if currency.BaseDenom != "" {
			if _, found := uniqueBaseDenomTracker[currency.BaseDenom]; found {
				utils.PrintlnStdErr("ERR: Duplicated base denom found:", currency.BaseDenom)
				return false, currency.BaseDenom
			}
			uniqueBaseDenomTracker[currency.BaseDenom] = true
		}

		if currency.DisplayDenom != "" {
			if _, found := uniqueDisplayDenomTracker[currency.DisplayDenom]; found {
				utils.PrintlnStdErr("ERR: Duplicated display denom found:", currency.DisplayDenom)
				return false, currency.DisplayDenom
			}
			uniqueDisplayDenomTracker[currency.DisplayDenom] = true
		}

		if currency.IbcRepresentation != "" {
			if _, found := uniqueIbcRepresentationTracker[currency.IbcRepresentation]; found {
				utils.PrintlnStdErr("ERR: Duplicated IBC representation found:", currency.IbcRepresentation)
				return false, currency.IbcRepresentation
			}
			uniqueIbcRepresentationTracker[currency.IbcRepresentation] = true
		}

	}
	if !foundMain {
		utils.PrintlnStdErr("ERR: At least one main currency is required")
		return false, ""
	}

	return true, ""
}

func isValidCurrency(currency valtypes.CurrencyChainDefinition, chainPath string) bool {
	if currency.DisplayDenom == "" {
		utils.PrintlnStdErr("ERR: Display denom is required")
		return false
	}
	if strings.TrimSpace(currency.DisplayDenom) != currency.DisplayDenom {
		utils.PrintlnStdErr("ERR: Display denom must not have leading or trailing spaces")
		return false
	}
	if strings.Contains(currency.DisplayDenom, "  ") {
		utils.PrintlnStdErr("ERR: Display denom must not have consecutive spaces")
		return false
	}
	if !regexp.MustCompile(`^[a-zA-Z\d\s-_]+$`).MatchString(currency.DisplayDenom) {
		utils.PrintlnStdErr("ERR: Display denom must be alphanumeric, space, underscore, or dash")
		return false
	}
	if currency.BaseDenom == "" {
		utils.PrintlnStdErr("ERR: Base denom is required")
		return false
	}
	if strings.TrimSpace(currency.BaseDenom) != currency.BaseDenom {
		utils.PrintlnStdErr("ERR: Base denom must not have leading or trailing spaces")
		return false
	}
	if strings.Contains(currency.BaseDenom, "  ") {
		utils.PrintlnStdErr("ERR: Base denom must not have consecutive spaces")
		return false
	}
	if strings.Contains(currency.BaseDenom, "//") {
		utils.PrintlnStdErr("ERR: Base denom must not have consecutive slashes")
		return false
	}
	if strings.Contains(currency.BaseDenom, "--") {
		utils.PrintlnStdErr("ERR: Base denom must not have consecutive dashes")
		return false
	}
	if strings.Contains(currency.BaseDenom, "__") {
		utils.PrintlnStdErr("ERR: Base denom must not have consecutive underscores")
		return false
	}
	if !regexp.MustCompile(`^[a-zA-Z\d\s-_/]+$`).MatchString(currency.BaseDenom) {
		utils.PrintlnStdErr("ERR: Base denom must be alphanumeric, space, underscore, dash, or slash")
		return false
	}
	if currency.IbcRepresentation != "" {
		if strings.TrimSpace(currency.IbcRepresentation) != currency.IbcRepresentation {
			utils.PrintlnStdErr("ERR: IBC representation must not have leading or trailing spaces")
			return false
		}
		if !regexp.MustCompile(`^ibc/[A-F\d]{64}$`).MatchString(currency.IbcRepresentation) {
			//goland:noinspection SpellCheckingInspection
			utils.PrintlnStdErr("ERR: IBC representation must match format ibc/32BYTESHASH")
			return false
		}
	}
	if currency.BridgeDenom != "" {
		if strings.TrimSpace(currency.BridgeDenom) != currency.BridgeDenom {
			utils.PrintlnStdErr("ERR: Bridge denom must not have leading or trailing spaces")
			return false
		}
		if strings.Contains(currency.BridgeDenom, "  ") {
			utils.PrintlnStdErr("ERR: Bridge denom must not have consecutive spaces")
			return false
		}
		if strings.Contains(currency.BridgeDenom, "//") {
			utils.PrintlnStdErr("ERR: Bridge denom must not have consecutive slashes")
			return false
		}
		if strings.Contains(currency.BridgeDenom, "--") {
			utils.PrintlnStdErr("ERR: Bridge denom must not have consecutive dashes")
			return false
		}
		if strings.Contains(currency.BridgeDenom, "__") {
			utils.PrintlnStdErr("ERR: Bridge denom must not have consecutive underscores")
			return false
		}
		if !regexp.MustCompile(`^[a-zA-Z\d\s-_/]+$`).MatchString(currency.BridgeDenom) {
			utils.PrintlnStdErr("ERR: Bridge denom must be alphanumeric, space, underscore, dash, or slash")
			return false
		}
	}
	if currency.Decimals < 0 {
		utils.PrintlnStdErr("ERR: Decimals must be non-negative")
		return false
	}
	if currency.Decimals > 18 {
		utils.PrintlnStdErr("ERR: Decimals must not exceed 18")
		return false
	}
	if !isValidLogo(currency.Logo, chainPath) {
		utils.PrintlnStdErr("ERR: Bad currency logo:", currency.Logo)
		return false
	}

	switch currency.Type {
	case "main":
		return true
	case "regular":
		return true
	default:
		utils.PrintlnStdErr("ERR: Not recognized currency type:", currency.Type)
		return false
	}
}

func isValidEvmHexChainId(cd valtypes.ChainDefinition) bool {
	if !regexp.MustCompile(`^0x[a-fA-F\d]+$`).MatchString(cd.EVM.ChainId) {
		utils.PrintlnStdErr("ERR: EVM hex chain id must be 0x followed by hexadecimal characters")
		return false
	}

	var checkWithCosmosChainId bool
	if cd.IsRollAppChain() {
		checkWithCosmosChainId = true
	} else if regexp.MustCompile(`^[a-z\d]+_\d+-\d+$`).MatchString(cd.ChainId) {
		checkWithCosmosChainId = true
	}

	if checkWithCosmosChainId {
		spl := strings.Split(cd.ChainId, "_")
		if len(spl) != 2 {
			utils.PrintlnStdErr("ERR: EVM RollApp chain id must have format <alphanumeric>_<number>-<number>")
			return false
		}
		spl = strings.Split(spl[1], "-")
		chainIdFromCosmos, err := strconv.ParseInt(spl[0], 10, 64)
		if err != nil {
			panic(err)
		}
		chainIdFromEvm, err := strconv.ParseInt(cd.EVM.ChainId, 0, 64)
		if err != nil {
			panic(err)
		}
		if chainIdFromCosmos != chainIdFromEvm {
			utils.PrintfStdErr("ERR: EVM hex chain id %d must match with the chain id from cosmos chain id %d\n", chainIdFromEvm, chainIdFromCosmos)
			return false
		}
	}

	return true
}

func isValidDA(cd valtypes.ChainDefinition) bool {
	if !cd.IsRollAppChain() {
		if cd.DA != "" {
			utils.PrintlnStdErr("ERR: DA must be empty for non-RollApp chains")
			return false
		}
		return true
	}
	if cd.IsRollAppChain() && cd.DA == "" {
		utils.PrintlnStdErr("ERR: DA is required for RollApp chains")
		return false
	}
	switch cd.DA {
	case "Avail":
		return true
	case "Celestia":
		return true
	case "local":
		return true
	default:
		utils.PrintlnStdErr("ERR: DA must be one of: 'Avail', 'Celestia', 'local'")
		return false
	}
}

func isValidOptionalWebsiteUrl(websiteUrl string) bool {
	if websiteUrl == "" {
		return true
	}

	if strings.TrimSpace(websiteUrl) != websiteUrl {
		utils.PrintlnStdErr("ERR: url must not have leading or trailing spaces")
		return false
	}

	if strings.Contains(websiteUrl, " ") {
		utils.PrintlnStdErr("ERR: url must not contains space")
		return false
	}

	return true
}

func isValidBech32Prefix(bech32Prefix string) bool {
	if bech32Prefix == "" {
		utils.PrintlnStdErr("ERR: bech32 prefix can not be empty")
		return false
	}
	if strings.TrimSpace(bech32Prefix) != bech32Prefix {
		utils.PrintlnStdErr("ERR: bech32 prefix must not have leading or trailing spaces")
		return false
	}
	if strings.ToLower(bech32Prefix) != bech32Prefix {
		utils.PrintlnStdErr("ERR: bech32 prefix must be lowercase")
		return false
	}
	if strings.Contains(bech32Prefix, " ") {
		utils.PrintlnStdErr("ERR: bech32 prefix must not contains space")
		return false
	}
	if strings.Contains(bech32Prefix, "1") {
		utils.PrintlnStdErr("ERR: bech32 prefix must not contains '1'")
		return false
	}
	if !regexp.MustCompile(`^[a-z\d]+$`).MatchString(bech32Prefix) {
		utils.PrintlnStdErr("ERR: bech32 prefix must be lowercase alphanumeric")
		return false
	}
	return true
}

func isValidUrls(urls []string) bool {
	if len(urls) == 1 && urls[0] == "" {
		return true
	}
	for _, url := range urls {
		if !isValidUrl(url) {
			return false
		}
	}
	return true
}

func isValidUrl(url string) bool {
	if url == "" {
		utils.PrintlnStdErr("ERR: url can not be empty")
		return false
	}
	if strings.TrimSpace(url) != url {
		utils.PrintlnStdErr("ERR: url must not have leading or trailing spaces")
		return false
	}
	if strings.Contains(url, " ") {
		utils.PrintlnStdErr("ERR: url must not contains space")
		return false
	}
	return true
}

func isValidChainName(chainName string) bool {
	if chainName == "" {
		utils.PrintlnStdErr("ERR: chain name can not be empty")
		return false
	}
	if strings.TrimSpace(chainName) != chainName {
		utils.PrintlnStdErr("ERR: chain name must not have leading or trailing spaces")
		return false
	}
	if strings.Contains(chainName, "  ") {
		utils.PrintlnStdErr("ERR: chain name must not have consecutive spaces")
		return false
	}
	if regexp.MustCompile(`[<>/\\%]`).MatchString(chainName) {
		// < > to prevent xss
		// / \ % to prevent path traversal and conflict
		utils.PrintlnStdErr("ERR: chain name contains prohibited characters: <, >, /, \\, %")
		return false
	}
	return true
}

func isValidChainId(chainId string, isEvmRollApp bool) bool {
	if chainId == "" {
		utils.PrintlnStdErr("ERR: chain id can not be empty")
		return false
	}
	if len(chainId) < 3 {
		utils.PrintlnStdErr("ERR: chain id is too short")
		return false
	}
	if strings.Contains(chainId, "--") {
		utils.PrintlnStdErr("ERR: chain id must not have consecutive dashes")
		return false
	}
	if strings.Contains(chainId, "__") {
		utils.PrintlnStdErr("ERR: chain id must not have consecutive underscores")
		return false
	}
	if strings.ToLower(chainId) != chainId {
		utils.PrintlnStdErr("ERR: chain id must be lowercase")
		return false
	}
	firstChar := chainId[0]
	if firstChar < 'a' || firstChar > 'z' {
		utils.PrintlnStdErr("ERR: chain id must start with a letter")
		return false
	}
	if isEvmRollApp {
		valid := regexp.MustCompile(`^[a-z\d]+_\d+-\d+$`).MatchString(chainId)
		if !valid {
			utils.PrintlnStdErr("ERR: chain id not match for EVM RollApp")
		}
		return valid
	}
	if regexp.MustCompile(`^[a-z\d]+$`).MatchString(chainId) {
		// only alphanumeric
		return true
	}
	if regexp.MustCompile(`^[a-z\d]+-\d+$`).MatchString(chainId) {
		// cosmos chain id
		return true
	}
	if regexp.MustCompile(`^[a-z\d]+_\d+-\d+$`).MatchString(chainId) {
		// EVM chain id
		return true
	}
	if regexp.MustCompile(`^[a-z\d\-]+-[a-z\d]+$`).MatchString(chainId) {
		// multiple dash chain id
		return true
	}

	return false
}
