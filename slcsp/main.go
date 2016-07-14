package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
)

type RateArea struct {
	State string
	Num   string
}

var RateAreas = make(map[string]map[RateArea]struct{}, 0)
var Plans = make(map[RateArea][]float64, 0)

func getSLCSP(zip string) []byte {
	// If a zipcode has more than one rate area, it's impossible to determine
	// the SLCSP from the zipcode alone, as each rate area has a unique SLCSP.
	if len(RateAreas[zip]) != 1 {
		return nil
	}

	// Get rate area for the zipcode. This will only loop once.
	rateArea := RateArea{}
	for ra, _ := range RateAreas[zip] {
		rateArea = ra
	}

	// Impossible to determine SLCSP if the zipcode's rate area has no silver
	// plans.
	if len(Plans[rateArea]) == 0 {
		return nil
	}

	var (
		rate bytes.Buffer

		min float64 // FIRST lowest cost silver plan rate
		mid float64 // SECOND lowest cost silver plan rate
	)
	for _, rate := range Plans[rateArea] {
		if rate < min || min == 0 {
			min = rate
		}
		if (rate < mid && rate > min) || mid == 0 {
			mid = rate
		}
	}

	rate.WriteString(fmt.Sprintf("%.2f", mid))

	return rate.Bytes()
}

// Load and populate RateAreas and Plans from plans.csv and zips.csv.
func init() {
	plansFile, err := os.Open("plans.csv")
	if err != nil {
		panic(err)
	}
	plans := csv.NewReader(plansFile)

	// Discard the header line.
	plans.Read()

	var (
		record   []string
		rate     float64
		rateArea = RateArea{}
	)
	for {
		if record, err = plans.Read(); err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		// Only silver plans are used to determine SLCSPs.
		if record[2] != "Silver" {
			continue
		}

		// Get the rate area this silver plan belongs to.
		rateArea = RateArea{
			State: record[1],
			Num:   record[4],
		}

		// Get the rate of this silver plan.
		if rate, err = strconv.ParseFloat(record[3], 8); err != nil {
			panic(err)
		}

		// Add this silver plan to the rate area's collection.
		Plans[rateArea] = append(
			Plans[rateArea],
			rate,
		)

		rateArea = RateArea{}
	}

	zipsFile, err := os.Open("zips.csv")
	if err != nil {
		panic(err)
	}
	zips := csv.NewReader(zipsFile)

	// Discard the header line.
	zips.Read()

	for {
		if record, err = zips.Read(); err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}

		// Get a rate area of this zip.
		rateArea = RateArea{
			State: record[1],
			Num:   record[4],
		}

		if _, ok := RateAreas[record[0]]; !ok {
			RateAreas[record[0]] = make(map[RateArea]struct{}, 0)
		}

		// Update the zipcode's list of rate areas.
		RateAreas[record[0]][rateArea] = struct{}{}

		rateArea = RateArea{}
	}
}

func main() {
	slcspFile, err := os.Open("slcsp.csv")
	if err != nil {
		panic(err)
	}
	slcsp := csv.NewReader(slcspFile)

	outFile, err := os.Create("slcsp_full.csv")
	if err != nil {
		panic(err)
	}

	var (
		row    bytes.Buffer
		record []string
		i      int
		offset int64 // current byte position in the file
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

		// If current line is the header, don't attempt to get the SLCSP.
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
