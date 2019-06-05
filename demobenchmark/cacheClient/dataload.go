package cacheClient

import (
	"encoding/csv"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
)

var haveNextNextGaussian bool
var nextNextGaussian float64

func generateKey(index int) string {
	return "not yet"
}

func GenerateValue(index int) BookData {
	csvFile, err := os.Open("./cacheClient/data.csv")
	if err != nil {
		panic(err)
	}
	defer csvFile.Close()

	csvReader := csv.NewReader(csvFile)

	var data BookData
	for i := 0; i <= index; i++ {
		row, err := csvReader.Read()
		if err == io.EOF {
			log.Fatalln(err)
			break
		} else if err != nil {
			panic(err) // or handle it another way
		}

		if i == index {
			data = BookData{
				ISBN:       row[0],
				BookTitle:  row[1],
				BookAuthor: row[2],
				Year:       row[3],
				Publisher:  row[4],
				ImageM:     row[6],
			}
			log.Println(data)
		}
	}

	return data
}

func GenerateIndex(min, max, variance, average int) int {

	var n int
	for {
		tmp := NextGaussian()
		n = int(float64(variance)*tmp + float64(average))
		if n >= min && max <= max {
			break
		}
	}

	return n
}

func NextGaussian() float64 {
	//Box-Muller method to generate values from normal distribution
	if haveNextNextGaussian {
		haveNextNextGaussian = false
		return nextNextGaussian
	} else {
		var v1, v2, s float64
		for {
			v1 = 2*rand.Float64() - 1// [-1,1)
			v2 = 2*rand.Float64() - 1
			s = v1*v1 + v2*v2
			if s < 1 && s != 0.0 {
				break
			}
		}
		var tmp2 = -2 * math.Log(s) / s
		var multiplier = math.Sqrt(tmp2)
		nextNextGaussian = v2 * multiplier
		haveNextNextGaussian = true
		return v1 * multiplier
	}
}
