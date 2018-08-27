package main

const (
	NUMBER_AT_PEG = 7.5
)

type IssuingCoinsRuledRangeItem struct {
	Min float64
	Max float64
	NumOfCoins float64
	IssuingCoinsWithinSec int64
}

type ContractingCoinsRuledRangeItem struct {
	Min float64
	Max float64
	Tax float64
	NumOfMiningBonds float64
}

func initIssuingCoinsRuledRanges() []*IssuingCoinsRuledRangeItem {
	return []*IssuingCoinsRuledRangeItem{
		&IssuingCoinsRuledRangeItem{
			Min: 1.0,
			Max: 1.1,
			NumOfCoins: NUMBER_AT_PEG,
			IssuingCoinsWithinSec: 5400, // 90 * 60
		},
		&IssuingCoinsRuledRangeItem{
			Min: 1.1,
			Max: 1.2,
			NumOfCoins: 15,
			IssuingCoinsWithinSec: 3600, // 60 * 60
		},
		&IssuingCoinsRuledRangeItem{
			Min: 1.2,
			Max: 1.3,
			NumOfCoins: 30,
			IssuingCoinsWithinSec: 2700, // 45 * 60
		},
		&IssuingCoinsRuledRangeItem{
			Min: 1.3,
			Max: 1.4,
			NumOfCoins: 60,
			IssuingCoinsWithinSec: 1800, // 30 * 60
		},
		&IssuingCoinsRuledRangeItem{
			Min: 1.4,
			Max: 1.5,
			NumOfCoins: 120,
			IssuingCoinsWithinSec: 900, // 15 * 60
		},
		&IssuingCoinsRuledRangeItem{
			Min: 1.5,
			Max: 999999,
			NumOfCoins: 240,
			IssuingCoinsWithinSec: 600, // 10 * 60
		},
	}
}

func initContractingCoinsRuledRanges() []*ContractingCoinsRuledRangeItem {
	return []*ContractingCoinsRuledRangeItem{
		&ContractingCoinsRuledRangeItem{
			Min: 0.9,
			Max: 1.0,
			Tax: 25,
			NumOfMiningBonds: NUMBER_AT_PEG,
		},
		&ContractingCoinsRuledRangeItem{
			Min: 0.8,
			Max: 0.9,
			Tax: 50,
			NumOfMiningBonds: NUMBER_AT_PEG,
		},
		&ContractingCoinsRuledRangeItem{
			Min: 0.7,
			Max: 0.8,
			Tax: 75,
			NumOfMiningBonds: NUMBER_AT_PEG,
		},
		&ContractingCoinsRuledRangeItem{
			Min: 0.6,
			Max: 0.7,
			Tax: 100,
			NumOfMiningBonds: NUMBER_AT_PEG,
		},
		&ContractingCoinsRuledRangeItem{
			Min: 0.0,
			Max: 0.6,
			Tax: 100,
			NumOfMiningBonds: NUMBER_AT_PEG,
		},
	}
}


func getIssuingCoinsRuledRangesItem(
	exchangeRate float64,
	issuingCoinsRuledRanges []*IssuingCoinsRuledRangeItem,
) (
	*IssuingCoinsRuledRangeItem,
)  {
	for _, rangeItem := range issuingCoinsRuledRanges {
		if exchangeRate >= rangeItem.Min && exchangeRate < rangeItem.Max {
			return rangeItem
		}
	}
	return nil
}


func getContractingCoinsRuledRangeItem(
	exchangeRate float64,
	contractingCoinsRuledRanges []*ContractingCoinsRuledRangeItem,
) (
	*ContractingCoinsRuledRangeItem,
)  {
	for _, rangeItem := range contractingCoinsRuledRanges {
		if exchangeRate >= rangeItem.Min && exchangeRate < rangeItem.Max {
			return rangeItem
		}
	}
	return nil
}