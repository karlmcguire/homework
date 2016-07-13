package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
)

type TransactionLog struct {
	DebitTotal      float64
	CreditTotal     float64
	AutopaysStarted int
	AutopaysStopped int
	Users           map[uint64]float64
}

const (
	T_DEBIT = iota
	T_CREDIT
	T_START
	T_END
	T_INVALID
)

func toNumber(slice []byte, out interface{}) error {
	return binary.Read(
		bytes.NewReader(slice),
		binary.BigEndian,
		out,
	)
}

func NewTransactionLog(binaryData []byte, recordCount int) *TransactionLog {
	var (
		err error

		currentUserId uint64
		currentType   byte
		currentAmount float64

		binaryIndex int
		recordIndex int

		transactionLog = &TransactionLog{
			Users: make(map[uint64]float64, 0),
		}
	)

	for recordIndex = 0; recordIndex < recordCount; recordIndex++ {
		currentType = binaryData[binaryIndex]
		if uint8(currentType) >= T_INVALID {
			panic(fmt.Errorf("invalid type at record %d\n", recordIndex))
		}

		if err = toNumber(
			binaryData[binaryIndex+5:binaryIndex+13],
			&currentUserId,
		); err != nil {
			panic(fmt.Errorf("invalid binary at record %d\n", recordIndex))
		}

		if currentType == T_CREDIT || currentType == T_DEBIT {
			if err = toNumber(
				binaryData[binaryIndex+13:binaryIndex+21],
				&currentAmount,
			); err != nil {
				panic(fmt.Errorf("invalid binary at record %d\n", recordIndex))
			}
			binaryIndex += 21
		} else {
			binaryIndex += 13
		}

		switch currentType {
		case T_CREDIT:
			transactionLog.Users[currentUserId] -= currentAmount
			transactionLog.CreditTotal += currentAmount
		case T_DEBIT:
			transactionLog.Users[currentUserId] += currentAmount
			transactionLog.DebitTotal += currentAmount
		case T_START:
			transactionLog.AutopaysStarted++
		case T_END:
			transactionLog.AutopaysStopped++
		}
	}
	return transactionLog
}

func main() {
	binaryData, err := ioutil.ReadFile("txnlog.dat")
	if err != nil {
		panic(err)
	}
	if len(binaryData) < 9 {
		panic("invalid header")
	}
	if string(binaryData[:4]) != "MPS7" {
		panic("magic string != MPS7")
	}
	if binaryData[4] != byte(0x01) {
		panic("this program was built for version 1")
	}

	var recordCount uint32
	if err = toNumber(binaryData[5:9], &recordCount); err != nil {
		panic("invalid record count")
	}
	transactionLog := NewTransactionLog(binaryData[9:], int(recordCount))

	fmt.Printf(
		"$%.2f total dollars debited\n$%.2f total dollars credited\n\n",
		transactionLog.DebitTotal,
		transactionLog.CreditTotal,
	)

	fmt.Printf(
		"%d autopays were started\n%d autopays were stopped\n\n",
		transactionLog.AutopaysStarted,
		transactionLog.AutopaysStopped,
	)

	fmt.Printf(
		"balance of user 2456938384156277127:\n\t$%.2f\n",
		transactionLog.Users[2456938384156277127],
	)
}
