package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/jung-kurt/gofpdf"
)

type trombConf struct {
	Trombs []tromb `json:"trombs"`
}

type config struct {
	neededSum int64
	trombFile string
}

type tromb struct {
	Name  string `json:"name"`
	Value int64  `json:"value"`
	Rest  int64  `json:"-"`
}

type kombinacije struct {
	suma  int64
	kombo []tromb
}

var c config

func main() {

	flag.Int64Var(&c.neededSum, "expected", 100, "needed sum")
	flag.StringVar(&c.trombFile, "file", "", "file with tromb values")

	flag.Parse()

	confData, err := readConfFile(c.trombFile)
	if err != nil {
		log.Printf("Error: %s", err.Error())
		return
	}

	var trombConfig trombConf
	err = json.Unmarshal(confData, &trombConfig)
	if err != nil {
		log.Print(err.Error())
		return
	}

	tmp := generateKomb(trombConfig.Trombs)
	komb := forPdf(tmp)

	err = createPDF(komb)
	if err != nil {
		log.Print(err.Error())
		return
	}

}

func generateKomb(t []tromb) [][]tromb {
	komb := [][]tromb{[]tromb{}}
	l := len(t)
	start := 0

	for start < l {
		tmpL := len(komb)
		kombStart := 0

		for kombStart < tmpL {

			arr := make([]tromb, len(komb[kombStart]))
			copy(arr, komb[kombStart])

			arr = append(arr, t[start])
			komb = append(komb, arr)

			kombStart++
		}

		start++
	}

	return komb
}

func forPdf(t [][]tromb) []kombinacije {
	komb := []kombinacije{}

	for _, v := range t[1:] {
		k := kombinacije{}

		for _, tr := range v {
			tr.Rest = tr.Value
			nextSum := tr.Value + k.suma
			rest := nextSum - c.neededSum

			if nextSum > c.neededSum {
				k.suma = (tr.Value - rest) + k.suma
				tr.Rest = rest
			} else {
				k.suma = nextSum
				tr.Rest = tr.Rest - tr.Value
			}
			k.kombo = append(k.kombo, tr)
		}

		komb = append(komb, k)
	}

	return komb
}

func sum(values []int64) int64 {
	sum := int64(0)

	for _, v := range values {
		sum = v + sum
	}

	return sum
}

func readConfFile(path string) ([]byte, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func createPDF(k []kombinacije) error {

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 10)

	pdf.Cell(10, 10, fmt.Sprintf("Expected (best value): %v", c.neededSum))
	pdf.Ln(-1)
	pdf.Cell(10, 10, "Unit of measurement: m (meter)")
	pdf.Ln(-1)
	pdf.Line(pdf.GetX(), pdf.GetY(), pdf.GetX()+150, pdf.GetY())

	for _, val := range k {
		pdf.Cell(10, 10, fmt.Sprintf("Spent: %v", val.suma))

		pdf.Ln(-1)

		for _, t := range val.kombo {
			pdf.Cell(10, 10, fmt.Sprintf("Tromba: %s, Original value: %v, Rest: %v", t.Name, t.Value, t.Rest))

			pdf.Ln(-1)
		}

		pdf.Line(pdf.GetX(), pdf.GetY(), pdf.GetX()+150, pdf.GetY())
	}

	return pdf.OutputFileAndClose("calc.pdf")
}
