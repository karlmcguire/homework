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

// getNumber is a helper function for converting binary big endian numbers to
// the golang equivalent.
func getNumber(slice []byte, out interface{}) error {
	return binary.Read(bytes.NewReader(slice), binary.BigEndian, out)
}

// NewTransactionLog returns a TransactionLog after parsing a MPS7 transaction
// log byte slice, b, with c being the total amount of records.
func NewTransactionLog(b []byte, c int) *TransactionLog {
	var (
		transactionLog = &TransactionLog{
			Users: make(map[uint64]float64, 0),
		}

		iden   byte    // current record identifier
		user   uint64  // current record userId
		amount float64 // current record amount (used if CREDIT || DEBIT)

		i int // current byte index, i < len(b)
		r int // current record index, r < c

		err error
	)

	for r = 0; r < c; r++ {
		// Determine current record type.
		if iden = b[i]; iden > END {
			panic(fmt.Errorf("invalid type at record %d\n", r))
		}

		// Determine if we need to get the userId and amount.
		if iden == DEBIT || iden == CREDIT {
			// Attempt to get the userId of current record.
			if err = getNumber(b[i+5:i+13], &user); err != nil {
				panic(fmt.Errorf("invalid userId binary at record %d\n", r))
			}

			// Attempt to get the amount of current record.
			if err = getNumber(b[i+13:i+21], &amount); err != nil {
				panic(fmt.Errorf("invalid amount binary at record %d\n", r))
			}

			if iden == DEBIT {
				// Add amount to user's balance and debit total.
				transactionLog.Users[user] += amount
				transactionLog.DebitTotal += amount
			} else {
				// Subtract amount from user's balance, add to credit total.
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

	// Determine the total amount of records contained in the file.
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
