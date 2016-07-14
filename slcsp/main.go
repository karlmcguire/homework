package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
)

var RateAreas = make(map[string][]*RateArea, 0)

type RateArea struct {
	State string
	Num   int
}

func (r *RateArea) String() string {
	return fmt.Sprintf("%s%d", r.State, r.Num)
}

var Plans = make(map[string][]*Plan, 0)

type Plan struct {
	Metal string
	Rate  float64
}

func setPlan(metal, rate string, out *Plan) {
	out.Metal = metal

	var err error
	if out.Rate, err = strconv.ParseFloat(rate, 8); err != nil {
		panic(err)
	}
}

func setRateArea(state, num string, out *RateArea) {
	out.State = state

	var err error
	if out.Num, err = strconv.Atoi(num); err != nil {
		panic(err)
	}
}

func getSLCSP(zip string) []byte {
	var (
		min float64
		mid float64

		test []float64

		buf bytes.Buffer
	)
	for _, v := range RateAreas[zip] {
		for _, p := range Plans[v.String()] {
			if p.Metal != "Silver" {
				continue
			}
			if p.Rate < min || min == 0 {
				min = p.Rate
			}
			if (p.Rate < mid && p.Rate > min) || mid == 0 {
				mid = p.Rate
			}

			if zip == "61232" {
				test = append(test, p.Rate)
			}
		}

		if len(Plans[v.String()]) == 0 {
			fmt.Println(zip)
		}
	}

	if test != nil {
		sort.Float64s(test)

		fmt.Println(test)
	}

	if mid != 0.0 {
		buf.WriteString(fmt.Sprintf("%.2f", mid))
	}

	return buf.Bytes()
}

func init() {
	plansFile, err := os.Open("plans.csv")
	if err != nil {
		panic(err)
	}

	plans := csv.NewReader(plansFile)
	_, _ = plans.Read()

	var (
		plan     = &Plan{}
		rateArea = &RateArea{}
		record   []string
	)
	for {
		if record, err = plans.Read(); err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		setPlan(
			record[2], // metal
			record[3], // rate
			plan,
		)

		setRateArea(
			record[1], // state letters
			record[4], // number
			rateArea,
		)

		Plans[rateArea.String()] = append(Plans[rateArea.String()], plan)

		rateArea = &RateArea{}
		plan = &Plan{}
	}

	zipsFile, err := os.Open("zips.csv")
	if err != nil {
		panic(err)
	}

	zips := csv.NewReader(zipsFile)
	_, _ = zips.Read()

	for {
		if record, err = zips.Read(); err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		setRateArea(
			record[1], // state letters
			record[4], // number
			rateArea,
		)

		RateAreas[record[0]] = append(RateAreas[record[0]], rateArea)

		rateArea = &RateArea{}
	}
}

func main() {
	slcspFile, err := os.Open("slcsp.csv")
	if err != nil {
		panic(err)
	}
	slcsp := csv.NewReader(slcspFile)

	outFile, err := os.Create("out.csv")
	if err != nil {
		panic(err)
	}

	var (
		row    bytes.Buffer
		offset int64
		record []string
		i      int
	)
	for {
		if record, err = slcsp.Read(); err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		row.WriteString(record[0])
		row.WriteString(",")

		if i == 0 {
			row.WriteString(record[1])
		} else {
			row.Write(getSLCSP(record[0]))
		}

		row.WriteString("\n")

		outFile.WriteAt(row.Bytes(), offset)

		offset += int64(row.Len())
		row.Reset()
		i++
	}

	outFile.Sync()
}
