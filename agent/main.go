package main

import (
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"time"
	"runtime"
	"strconv"
	// "math"
)

const (
	AVERAGE_MINED_BLOCK_IN_SEC = 60 // 1 min

)

type RPCError struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

type RPCNumOfCoinsAndBondsResponse struct {
	Id int `json:"id"`
	RPCError *RPCError `json:"error"`
	Result map[string]float64 `json:"result"`
}

type RPCHashResponse struct {
	Id int `json:"id"`
	RPCError *RPCError `json:"error"`
	Result string `json:"result"`
}

func getNumOfCoinsAndBonds(
	client *HttpClient,
	method string,
	result *map[string]float64,
) (error) {
	var rpcResponse = &RPCNumOfCoinsAndBondsResponse{}
	err := client.RPCCall(method, nil, rpcResponse)
	if err != nil {
		return err
	}
	if rpcResponse.RPCError != nil {
		return fmt.Errorf("%s", rpcResponse.RPCError.Message)
	}
	fmt.Println("haha: ", rpcResponse.Result)
	*result = rpcResponse.Result
	return nil
}

func sendActionParamToBlockchainNode(
	client *HttpClient,
	method string,
	param map[string]interface{},
) (error) {
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


func process(
	issuingCoinsRuledRanges []*IssuingCoinsRuledRangeItem,
	contractingCoinsRuledRanges []*ContractingCoinsRuledRangeItem,
) error {
	// hardcoded eligibleAgentIDs here
	eligibleAgentIDs := []string{"agent_1", "agent_2", "agent_3"}

	client := NewHttpClient()
	var coinsAndBondsMap map[string]float64
	err := getNumOfCoinsAndBonds(client, "getNumberOfCoinsAndBonds", &coinsAndBondsMap)
	if err != nil {
		return err
	}
	numOfCoins := coinsAndBondsMap[TX_OUT_COIN_TYPE]
	numOfBonds := coinsAndBondsMap[TX_OUT_BOND_TYPE]

	// TODO: get exchange rate from external exchange
	exchangeRate := 0.65
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
			"agentId": getenv("AGENT_PUBLIC_KEY", "agent_1"),
			"agentSig": "", // sig here
			"numOfCoins": actualIssuingCoins,
			"numOfBonds": 0, // TODO: re-calculate this value
			"tax": 0,
			"eligibleAgentIDs": eligibleAgentIDs,
		}
		return sendActionParamToBlockchainNode(client, "createActionParamsTrasaction", param)
	}
	// price < peg
	contractingCoinsRuledRangeItem := getContractingCoinsRuledRangeItem(
		exchangeRate,
		contractingCoinsRuledRanges,
	)
	param := map[string]interface{}{
		"agentId": getenv("AGENT_PUBLIC_KEY", "agent_1"),
		"agentSig": "", // sig here
		"numOfCoins": 0, // TODO: re-calculate this value
		"numOfBonds": contractingCoinsRuledRangeItem.NumOfMiningBonds,
		"tax": contractingCoinsRuledRangeItem.Tax,
		"eligibleAgentIDs": eligibleAgentIDs,
	}

	// Hardcoded for demo
	// Agent 1

	// numCoins, _ := strconv.Atoi(getenv("NUM_COINS_FAKE", "120"))
	// param := map[string]interface{}{
	// 	"agentId": getenv("AGENT_PUBLIC_KEY", "agent_123456789"),
	// 	"numOfIssuingCoins": numCoins, // TODO: re-calculate this value
	// 	"numOfIssuingBonds": 0,
	// 	"tax": 0,
	// }

	// // Agent 2
	// param := map[string]interface{}{
	// 	"agentId": getenv("AGENT_PUBLIC_KEY", "agent_123456789"),
	// 	"numOfIssuingCoins": 130, // TODO: re-calculate this value
	// 	"numOfIssuingBonds": 0
	// 	"tax": 0,
	// }

	// // Agent 3
	// param := map[string]interface{}{
	// 	"agentId": getenv("AGENT_PUBLIC_KEY", "agent_123456789"),
	// 	"numOfIssuingCoins": 140, // TODO: re-calculate this value
	// 	"numOfIssuingBonds": 0
	// 	"tax": 0,
	// }

	return sendActionParamToBlockchainNode(client, "createActionParamsTrasaction", param)
}

func run() {
	// Agent re-calculates every 60s
	deplayTimeInSec, _ := strconv.Atoi(getenv("DELAY_TIME_IN_SEC", "60"))
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
	time.Sleep(2*time.Second)
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