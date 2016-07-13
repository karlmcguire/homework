package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
)

type TransactionLog struct {
	DebitBalance    float64
	CreditBalance   float64
	AutopaysStarted int
	AutopaysStopped int

	Users map[uint64]*User
}

type User struct {
	Balance float64
	Records []Record
}

type Record struct {
	Type      byte
	Timestamp uint32
	Amount    float64
}

const (
	T_DEBIT = iota
	T_CREDIT
	T_START
	T_END
)

func NewTransactionLog(binaryData []byte, recordCount int) *TransactionLog {
	var (
		err error

		transactionLog *TransactionLog = &TransactionLog{
			Users: make(map[uint64]*User, 0),
		}

		currentRecord Record
		currentUserId uint64

		binaryIndex int
		recordIndex int
	)

	for recordIndex = 0; recordIndex < recordCount; recordIndex++ {
		switch binaryData[binaryIndex] {
		case T_DEBIT:
			currentRecord.Type = T_DEBIT
		case T_CREDIT:
			currentRecord.Type = T_CREDIT
		case T_START:
			transactionLog.AutopaysStarted++
			currentRecord.Type = T_START
		case T_END:
			transactionLog.AutopaysStopped++
			currentRecord.Type = T_END
		default:
			goto typeErr
		}

		if err = binary.Read(
			bytes.NewReader(
				binaryData[binaryIndex+1:binaryIndex+5],
			),
			binary.BigEndian,
			&currentRecord.Timestamp,
		); err != nil {
			goto binaryErr
		}

		if err = binary.Read(
			bytes.NewReader(
				binaryData[binaryIndex+5:binaryIndex+13],
			),
			binary.BigEndian,
			&currentUserId,
		); err != nil {
			goto binaryErr
		}

		if currentRecord.Type == T_CREDIT || currentRecord.Type == T_DEBIT {
			if err = binary.Read(
				bytes.NewReader(
					binaryData[binaryIndex+13:binaryIndex+21],
				),
				binary.BigEndian,
				&currentRecord.Amount,
			); err != nil {
				goto binaryErr
			}

			binaryIndex += 21
		} else {
			binaryIndex += 13
		}

		if _, ok := transactionLog.Users[currentUserId]; !ok {
			transactionLog.Users[currentUserId] = &User{
				Records: make([]Record, 0),
			}
		}

		if currentRecord.Type == T_CREDIT {
			transactionLog.Users[currentUserId].Balance += currentRecord.Amount
			transactionLog.CreditBalance += currentRecord.Amount
		} else if currentRecord.Type == T_DEBIT {
			transactionLog.Users[currentUserId].Balance -= currentRecord.Amount
			transactionLog.DebitBalance += currentRecord.Amount
		}

		transactionLog.Users[currentUserId].Records = append(
			transactionLog.Users[currentUserId].Records,
			currentRecord,
		)

		currentUserId, currentRecord = 0, Record{}
	}
	return transactionLog

typeErr:
	panic(fmt.Errorf("invalid record type at record %d\n", recordIndex))
	return nil

binaryErr:
	panic(fmt.Errorf("invalid binary at record %d\n", recordIndex))
	return nil
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
		panic("need version 1")
	}

	var recordCount uint32
	if err = binary.Read(
		bytes.NewReader(
			binaryData[5:9],
		),
		binary.BigEndian,
		&recordCount,
	); err != nil {
		panic(err)
	}

	transactionLog := NewTransactionLog(binaryData[9:], int(recordCount))
	fmt.Println(transactionLog.Users[2456938384156277127].Balance)
	fmt.Println(transactionLog.Users[2456938384156277127].Records)
	fmt.Println(transactionLog.CreditBalance)
	fmt.Println(transactionLog.DebitBalance)
	fmt.Println(
		transactionLog.AutopaysStarted,
		transactionLog.AutopaysStopped,
	)
}
