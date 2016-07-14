package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
)

var Zips = make(map[string][]*County, 0)

type County struct {
	State    string
	Fips     int
	Name     string
	RateArea int
}

var Plans = make(map[string]*Plan, 0)

type Plan struct {
	State    string
	Metal    string
	Rate     float64
	RateArea int
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
		record []string
	)
	for {
		if record, err = plans.Read(); err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		plan.State = record[1]
		plan.Metal = record[2]

		if plan.Rate, err = strconv.ParseFloat(record[3], 8); err != nil {
			panic(err)
		}

		if plan.RateArea, err = strconv.Atoi(record[4]); err != nil {
			panic(err)
		}

		Plans[record[0]] = plan

		plan = &Plan{}
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
		row    int
		offset int64
	)
	for {
		if record, err = slcsp.Read(); err != nil {
			if err == io.EOF {
				break
			}
		}

		outFile.WriteAt([]byte(record[0]), offset)
		offset += int64(len(record[0]))
		outFile.WriteAt([]byte(","), offset)
		offset += 1

		if row == 0 {
			outFile.WriteAt([]byte(record[1]), offset)
			offset += int64(len(record[1]))
			outFile.WriteAt([]byte("\n"), offset)
			offset += 1
		} else {
			outFile.WriteAt([]byte("\n"), offset)
			offset += 1
		}

		row++
	}

	outFile.Sync()

	fmt.Println(slcsp.Read())
	fmt.Println(len(Zips), len(Plans))

}

func main() {

}
