package main

import (
	"encoding/csv"
	"fmt"
	"os"
)

func init() {
	slcspFile, err := os.Open("slcsp.csv")
	if err != nil {
		panic(err)
	}
	slcsp := csv.NewReader(slcspFile)

	plansFile, err := os.Open("plans.csv")
	if err != nil {
		panic(err)
	}
	plans := csv.NewReader(plansFile)

	zipsFile, err := os.Open("zips.csv")
	if err != nil {
		panic(err)
	}
	zips := csv.NewReader(zipsFile)

	fmt.Println(slcsp.Read())
	fmt.Println(plans.Read())
	fmt.Println(zips.Read())
}

func main() {

}
