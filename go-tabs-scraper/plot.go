package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/whilei/go-tabs-scraper/lib"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func main() {
	datadir := flag.String("datadir", "data", "Root data directory. Will be created if not existing.")
	flag.Parse()

	matches, err := filepath.Glob(filepath.Join(*datadir, "block_*"))
	if err != nil {
		log.Fatalln(err)
	}

	width := vg.Length(800)
	w := width / vg.Length(len(matches))
	p, _ := plot.New()

	for mi, m := range matches {
		log.Printf("Reading match %d/%d %s\n", mi, len(matches), m)

		b, err := ioutil.ReadFile(m)
		if err != nil {
			log.Fatalln(err)
		}

		ap := &lib.AppBlock{}
		err = json.Unmarshal(b, ap)
		if err != nil {
			log.Fatalln(err)
		}

		// I want all txes
		// to compose a block

		minerBal, _ := lib.PrettyBalance(ap.MinerBalanceAtParent).Float64()

		group := plotter.Values{minerBal}
		lastBarChart, _ := plotter.NewBarChart(group, w)
		lastBarChart.Color = mustAddressToColor(ap.Header.Coinbase)
		lastBarChart.XMin = 0 + float64(width)/float64(len(matches))*float64(mi)

		p.Add(lastBarChart)
	}

	os.MkdirAll("out", os.ModePerm)
	if err := p.Save(800, 600, "out/stacked_balances.png"); err != nil {
		log.Panic(err)
	}
}

func mustAddressToColor(addr common.Address) color.Color {
	hex := strings.TrimPrefix(addr.Hex(), "0x")
	c, err := ParseHexColor(hex)
	if err != nil {
		panic(err)
	}
	return c
}

func ParseHexColor(s string) (c color.RGBA, err error) {

	// Prepend the hash character if it does not already exist.
	if !strings.HasPrefix(s, "#") {
		s = "#" + s
	}

	// Trim the string to a max length of 7.
	if len(s) > 7 {
		s = s[:7]
	}

	c.A = 0xff
	switch len(s) {
	case 7:
		_, err = fmt.Sscanf(s, "#%02x%02x%02x", &c.R, &c.G, &c.B)
	case 4:
		_, err = fmt.Sscanf(s, "#%1x%1x%1x", &c.R, &c.G, &c.B)
		// Double the hex digits:
		c.R *= 17
		c.G *= 17
		c.B *= 17
	default:
		err = fmt.Errorf("invalid length, must be 7 or 4, got: %v", len(s))

	}
	return
}
