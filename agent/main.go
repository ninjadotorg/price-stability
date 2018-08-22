package main

import (
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"time"
	"runtime"
	"strconv"
	"math"
)

type RPCError struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

type RPCNumberResponse struct {
	Id int `json:"id"`
	RPCError *RPCError `json:"error"`
	Result int `json:"result"`
}

type RPCHashResponse struct {
	Id int `json:"id"`
	RPCError *RPCError `json:"error"`
	Result string `json:"result"`
}

func getResultAsNumber(
	client *HttpClient,
	method string,
	result *int,
) (error) {
	var rpcResponse = &RPCNumberResponse{}
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

func process() error {
	client := NewHttpClient()
	var numOfCoins int
	err := getResultAsNumber(client, "getNumberOfCoins", &numOfCoins)
	if err != nil {
		return err
	}

	var numOfBonds int
	err = getResultAsNumber(client, "getNumberOfCoins", &numOfBonds)
	if err != nil {
		return err
	}

	// TODO: get exchange rate from external exchange
	exchangeRate := 0.8

	demand := exchangeRate * float64(numOfCoins)
	diff := demand - float64(numOfCoins)
	if diff > 0 { // price > peg, issuing more coins
		autoNeededIssuingCoins := diff - float64(numOfBonds)
		if autoNeededIssuingCoins > 0 {
			// TODO: manually issuing coins (= numOfBonds) to open auction market
			// make api call to add action param transaction via RPC
			param := map[string]interface{}{
				"agentId": getenv("AGENT_PUBLIC_KEY", "agent_123456789"),
				"numOfIssuingCoins": int(diff), // TODO: re-calculate this value
				"numOfIssuingBonds": 0, // TODO: re-calculate this value
				"tax": 0,
			}
			return sendActionParamToBlockchainNode(client, "createActionParamsTrasaction", param)
		}
		fmt.Println("Do nothing")
		return nil
	}
	// price < peg
	param := map[string]interface{}{
		"agentId": getenv("AGENT_PUBLIC_KEY", "agent_123456789"),
		"numOfIssuingCoins": 0, // TODO: re-calculate this value
		"numOfIssuingBonds": int(math.Ceil(math.Abs(diff) * 0.1)), // TODO: re-calculate this value
		"tax": 10,
	}
	return sendActionParamToBlockchainNode(client, "createActionParamsTrasaction", param)
}

func run() {
	// Agent re-calculates every 60s
	deplayTimeInSec, _ := strconv.Atoi(getenv("DELAY_TIME_IN_SEC", "20"))
	for {
		fmt.Println("Hello there again!!!")
		err := process()
		// TODO: log error here
		if err != nil {
			fmt.Println(err)
		}

		time.Sleep(time.Duration(deplayTimeInSec) * time.Second)
	}
}

func clearUpBeforeTerminating() {
	// TODO: do cleaning up here, probably send message to a channel in run func to stop loops
	fmt.Println("Wait for 2 second to finish processing")
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