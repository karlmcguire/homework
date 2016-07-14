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

var Zips = make(map[string][]*County, 0)

type County struct {
	State    string
	Fips     int
	Name     string
	RateArea int
}

var Plans = make(map[Id][]*Plan, 0)

type Plan struct {
	Metal string
	Rate  float64
}

type Id struct {
	State    string
	RateArea int
}

func getSLCSP(zip string) []byte {
	var (
		min float64
		mid float64
		max float64

		test []float64

		buf bytes.Buffer
	)
	for _, v := range Zips[zip] {
		for _, p := range Plans[Id{v.State, v.RateArea}] {
			if p.Metal != "Silver" {
				continue
			}

			if p.Rate < min || min == 0 {
				min = p.Rate
			}
			if (p.Rate < mid && p.Rate > min) || mid == 0 {
				mid = p.Rate
			}
			if (p.Rate < max && p.Rate > mid) || max == 0 {
				max = p.Rate
			}

			if zip == "61232" {
				test = append(test, p.Rate)
			}
		}
	}

	sort.Float64s(test)

	fmt.Println(test)

	if mid != 0.0 {
		buf.WriteString(fmt.Sprintf("%.2f", mid))
	}

	return buf.Bytes()
}

func init() {
	var err error

	plansFile, err := os.Open("plans.csv")
	if err != nil {
		panic(err)
	}
	plans := csv.NewReader(plansFile)
	_, _ = plans.Read()

	var (
		plan   = &Plan{}
		id     = Id{}
		record []string
	)
	for {
		if record, err = plans.Read(); err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		id.State = record[1]
		plan.Metal = record[2]

		if plan.Rate, err = strconv.ParseFloat(record[3], 8); err != nil {
			panic(err)
		}

		if id.RateArea, err = strconv.Atoi(record[4]); err != nil {
			panic(err)
		}

		Plans[id] = append(Plans[id], plan)

		plan = &Plan{}
		id = Id{}
	}

	zipsFile, err := os.Open("zips.csv")
	if err != nil {
		panic(err)
	}
	zips := csv.NewReader(zipsFile)
	_, _ = zips.Read()

	var (
		county = &County{}
	)
	for {
		if record, err = zips.Read(); err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		county.State = record[1]

		if county.Fips, err = strconv.Atoi(record[2]); err != nil {
			panic(err)
		}

		county.Name = record[3]

		if county.RateArea, err = strconv.Atoi(record[4]); err != nil {
			panic(err)
		}

		Zips[record[0]] = append(Zips[record[0]], county)

		county = &County{}
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
