package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"time"
)

type TransactionLog struct {
	DebitCount      int
	CreditCount     int
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
	S_TYPE = iota
	S_TIMESTAMP
	S_ID
	S_AMOUNT
	S_FINISH
)

const (
	T_DEBIT = iota
	T_CREDIT
	T_START
	T_END
)

func parse(data []byte) *TransactionLog {
	var (
		state      = S_TYPE
		amountFlag bool
		userId     uint64
		err        error
		i          int
		record     Record
	)

	tl := &TransactionLog{
		Users: make(map[uint64]*User, 0),
	}

	for i < len(data) {
		switch state {
		case S_TYPE:
			switch data[i] {
			case T_DEBIT:
				tl.DebitCount++
				record.Type = T_DEBIT
				amountFlag = true
			case T_CREDIT:
				tl.CreditCount++
				record.Type = T_CREDIT
				amountFlag = true
			case T_START:
				tl.AutopaysStarted++
				record.Type = T_START
				amountFlag = false
			case T_END:
				tl.AutopaysStopped++
				record.Type = T_END
				amountFlag = false
			default:
				goto typeErr
			}

			i++
			state = S_TIMESTAMP
		case S_TIMESTAMP:
			if err = binary.Read(
				bytes.NewReader(data[i:i+4]),
				binary.BigEndian,
				&record.Timestamp,
			); err != nil {
				goto binaryErr
			}

			i += 4
			state = S_ID
		case S_ID:
			if err = binary.Read(
				bytes.NewReader(data[i:i+8]),
				binary.BigEndian,
				&userId,
			); err != nil {
				goto binaryErr
			}

			i += 8
			if amountFlag {
				state = S_AMOUNT
			} else {
				state = S_FINISH
			}
		case S_AMOUNT:
			if err = binary.Read(
				bytes.NewReader(data[i:i+8]),
				binary.BigEndian,
				&record.Amount,
			); err != nil {
				goto binaryErr
			}

			i += 8
			state = S_FINISH
		case S_FINISH:
			if _, ok := tl.Users[userId]; !ok {
				tl.Users[userId] = &User{
					Records: make([]Record, 0),
				}
			}

			switch record.Type {
			case T_CREDIT:
				tl.Users[userId].Balance += record.Amount
			case T_DEBIT:
				tl.Users[userId].Balance -= record.Amount
			}

			tl.Users[userId].Records = append(
				tl.Users[userId].Records,
				record,
			)

			userId = 0
			record = Record{}
			state = S_TYPE
		}
	}

	return tl

typeErr:
	panic(fmt.Errorf("invalid record type at byte %d\n", i+9))
	return nil

binaryErr:
	panic(fmt.Errorf("invalid binary at byte %d\n", i+9))
	return nil
}

func main() {
	data, err := ioutil.ReadFile("txnlog.dat")
	if err != nil {
		panic(err)
	}
	if len(data) <= 9 {
		// todo
		panic("invalid header")
	}

	if string(data[:4]) != "MPS7" || data[4] != byte(0x01) {
		// todo
		panic("invalid header")
	}

	var recordCount uint32
	if err = binary.Read(
		bytes.NewReader(data[5:9]),
		binary.BigEndian,
		&recordCount,
	); err != nil {
		panic(err)
	}

	t1 := time.Now()
	tl := parse(data[9:])
	t1d := time.Since(t1)

	fmt.Println(t1d)

	fmt.Println(
		time.Unix(
			int64(
				tl.Users[2456938384156277127].Records[0].Timestamp,
			),
			0,
		),
	)
}
