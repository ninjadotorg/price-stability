package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/crypto/ed25519"
)

const (
	AVERAGE_MINED_BLOCK_IN_SEC = 60 // 1 min

	// these agent public keys should be pulled from agent management api
	AGENT1_PUBLIC_KEY_BASE64_ENCODED = "JtzdWgdm02rielHS5U9ZrdrJqj5Q9VhfmrfB/2X7POw="
	AGENT2_PUBLIC_KEY_BASE64_ENCODED = "Dzg9zPjznK69frGlWSiRN+9cD/Rwx1jr7tsAahDWxoY="
)

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type RPCNumOfCoinsAndBondsResponse struct {
	Id       int                `json:"id"`
	RPCError *RPCError          `json:"error"`
	Result   map[string]float64 `json:"result"`
}

type RPCHashResponse struct {
	Id       int       `json:"id"`
	RPCError *RPCError `json:"error"`
	Result   string    `json:"result"`
}

func getNumOfCoinsAndBonds(
	client *HttpClient,
	method string,
	result *map[string]float64,
) error {
	var rpcResponse = &RPCNumOfCoinsAndBondsResponse{}
	err := client.RPCCall(method, nil, rpcResponse)
	if err != nil {
		return err
	}
	if rpcResponse.RPCError != nil {
		return fmt.Errorf("%s", rpcResponse.RPCError.Message)
	}
	*result = rpcResponse.Result
	return nil
}

func sendActionParamToBlockchainNode(
	client *HttpClient,
	method string,
	param map[string]interface{},
) error {
	var rpcResponse = &RPCHashResponse{}
	err := client.RPCCall(method, []interface{}{param}, rpcResponse)
	if err != nil {
		return err
	}
	if rpcResponse.RPCError != nil {
		return fmt.Errorf("%s", rpcResponse.RPCError.Message)
	}
	return nil
}

func sign(message []byte) string {
	// public, private, _ := ed25519.GenerateKey(nil)
	// fmt.Println("public: ", (public))
	// fmt.Println("private: ", (private))

	privKeyStr := getenv("PRIVATE_KEY_BASE64_ENCODED", "")
	privKeyInBytes, _ := base64.StdEncoding.DecodeString(privKeyStr)

	signature := ed25519.Sign(privKeyInBytes, message)
	sigStr := base64.StdEncoding.EncodeToString(signature)
	return sigStr
}

func process(
	issuingCoinsRuledRanges []*IssuingCoinsRuledRangeItem,
	contractingCoinsRuledRanges []*ContractingCoinsRuledRangeItem,
) error {
	// hardcoded eligibleAgentIDs here
	eligibleAgentIDs := []string{AGENT1_PUBLIC_KEY_BASE64_ENCODED}

	client := NewHttpClient()
	var coinsAndBondsMap map[string]float64
	err := getNumOfCoinsAndBonds(client, "getNumberOfCoinsAndBonds", &coinsAndBondsMap)
	if err != nil {
		return err
	}
	numOfCoins := coinsAndBondsMap[TX_OUT_COIN_TYPE]
	numOfBonds := coinsAndBondsMap[TX_OUT_BOND_TYPE]

	// TODO: get exchange rate from external exchange
	exchangeRate := 1.25
	fmt.Println("exchangeRate: ", exchangeRate)

	demand := exchangeRate * numOfCoins
	diff := demand - numOfCoins
	if diff > 0 { // price > peg, issuing more coins
		autoNeededIssuingCoins := diff - numOfBonds
		if autoNeededIssuingCoins <= 0 {
			fmt.Println("Do nothing")
			return nil
		}

		// TODO: manually issuing coins (= numOfBonds) to open auction market

		issuingCoinsRuledRangesItem := getIssuingCoinsRuledRangesItem(
			exchangeRate,
			issuingCoinsRuledRanges,
		)

		numOfIssuingCoinsInNormalPace := (float64(issuingCoinsRuledRangesItem.IssuingCoinsWithinSec) * issuingCoinsRuledRangesItem.NumOfCoins) / AVERAGE_MINED_BLOCK_IN_SEC
		var actualIssuingCoins float64 = 0
		if numOfIssuingCoinsInNormalPace >= autoNeededIssuingCoins {
			actualIssuingCoins = issuingCoinsRuledRangesItem.NumOfCoins
		} else {
			actualIssuingCoins = (autoNeededIssuingCoins / numOfIssuingCoinsInNormalPace) * issuingCoinsRuledRangesItem.NumOfCoins
		}
		// make api call to add action param transaction via RPC
		param := map[string]interface{}{
			"agentId":          getenv("PUBLIC_KEY_BASE64_ENCODED", ""),
			"numOfCoins":       actualIssuingCoins,
			"numOfBonds":       0,
			"tax":              0,
			"eligibleAgentIDs": eligibleAgentIDs,
		}
		messageInBytes, _ := json.Marshal(param)
		param["agentSig"] = sign(messageInBytes)
		return sendActionParamToBlockchainNode(client, "createActionParamsTrasaction", param)
	}

	// price < peg
	contractingCoinsRuledRangeItem := getContractingCoinsRuledRangeItem(
		exchangeRate,
		contractingCoinsRuledRanges,
	)
	param := map[string]interface{}{
		"agentId":          getenv("PUBLIC_KEY_BASE64_ENCODED", ""),
		"numOfCoins":       0,
		"numOfBonds":       contractingCoinsRuledRangeItem.NumOfMiningBonds,
		"tax":              contractingCoinsRuledRangeItem.Tax,
		"eligibleAgentIDs": eligibleAgentIDs,
	}
	messageInBytes, _ := json.Marshal(param)
	param["agentSig"] = sign(messageInBytes)
	return sendActionParamToBlockchainNode(client, "createActionParamsTrasaction", param)
}

func run() {
	// Agent re-calculates every 60s
	deplayTimeInSec, _ := strconv.Atoi(getenv("DELAY_TIME_IN_SEC", "600"))
	issuingCoinsRuledRanges := initIssuingCoinsRuledRanges()
	contractingCoinsRuledRanges := initContractingCoinsRuledRanges()
	for {
		fmt.Println("Hello there again!!!")
		err := process(issuingCoinsRuledRanges, contractingCoinsRuledRanges)
		// TODO: log error here
		if err != nil {
			fmt.Println(err)
		}

		time.Sleep(time.Duration(deplayTimeInSec) * time.Second)
	}
}

func clearUpBeforeTerminating() {
	// TODO: do cleaning up here, probably send message to a channel in run func to stop loops
	fmt.Println("Wait for 2 seconds to finish processing")
	time.Sleep(2 * time.Second)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	go func() {
		sig := <-gracefulStop
		fmt.Printf("caught sig: %+v", sig)
		clearUpBeforeTerminating()
		os.Exit(0)
	}()
	run()
}
