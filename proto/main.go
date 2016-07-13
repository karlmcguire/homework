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
	DEBIT = iota
	CREDIT
	START
	END
)

func getNumber(slice []byte, out interface{}) error {
	return binary.Read(bytes.NewReader(slice), binary.BigEndian, out)
}

func NewTransactionLog(b []byte, c int) *TransactionLog {
	var (
		transactionLog = &TransactionLog{
			Users: make(map[uint64]float64, 0),
		}

		iden   byte
		user   uint64
		amount float64

		i int
		r int

		err error
	)

	for r = 0; r < c; r++ {
		iden = b[i]
		if iden > END {
			panic(fmt.Errorf("invalid type at record %d\n", r))
		}

		if iden == DEBIT || iden == CREDIT {
			if err = getNumber(b[i+5:i+13], &user); err != nil {
				panic(fmt.Errorf("invalid userId binary at record %d\n", r))
			}

			if err = getNumber(b[i+13:i+21], &amount); err != nil {
				panic(fmt.Errorf("invalid amount binary at record %d\n", r))
			}

			if iden == DEBIT {
				transactionLog.Users[user] += amount
				transactionLog.DebitTotal += amount
			} else {
				transactionLog.Users[user] -= amount
				transactionLog.CreditTotal += amount
			}

			i += 21
		} else {
			if iden == START {
				transactionLog.AutopaysStarted++
			} else {
				transactionLog.AutopaysStopped++
			}

			i += 13
		}
	}

	return transactionLog
}

func main() {
	b, err := ioutil.ReadFile("txnlog.dat")
	if err != nil {
		panic(err)
	}
	if len(b) < 9 {
		panic("incomplete header")
	}
	if string(b[:4]) != "MPS7" {
		panic("invalid header")
	}
	if b[4] != 0x01 {
		panic("version != 1")
	}

	var c uint32
	if err = getNumber(b[5:9], &c); err != nil {
		panic("invalid record count")
	}

	transactionLog := NewTransactionLog(b[9:], int(c))

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
