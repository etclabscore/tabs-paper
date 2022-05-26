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
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/montanaflynn/stats"
	"github.com/whilei/go-tabs-scraper/lib"
	"golang.org/x/image/colornames"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func main() {
	datadir := flag.String("datadir", "data", "Root data directory. Will be created if not existing.")
	maxSamples := flag.Int64("max-samples", 1000, "Max number of samples. (default=50)")
	flag.Parse()

	matches, err := filepath.Glob(filepath.Join(*datadir, "block_*"))
	if err != nil {
		log.Fatalln(err)
	}

	sampleSize := int64(len(matches))
	if *maxSamples < int64(len(matches)) {
		sampleSize = *maxSamples
	}

	width := vg.Length(800)
	w := width / vg.Length(sampleSize)

	p, _ := plot.New()
	p.Title.Text = fmt.Sprintf("Sampled Blocks: Miner (private) vs. Tx (public) balances, samples=%d", sampleSize)
	p.Y.Label.Text = "Summed Balance"
	p.X.Label.Text = "Blocks (discontinuous, but chronological)"
	p.Legend.Top = true

	tabSamples := []float64{}
	for mi, m := range matches {
		if int64(mi) >= sampleSize {
			break
		}
		log.Printf("Reading match %d/%d %s\n", mi, len(matches), m)

		tab := float64(0)

		b, err := ioutil.ReadFile(m)
		if err != nil {
			log.Fatalln(err)
		}

		ap := &lib.AppBlock{}
		err = json.Unmarshal(b, ap)
		if err != nil {
			log.Fatalln(err)
		}

		// Get the miner (private) balance
		minerBal, _ := lib.PrettyBalance(ap.MinerBalanceAtParent).Float64()

		// Add it to our TAB value / block
		tab += minerBal

		group := plotter.Values{minerBal}
		lastBarChart, _ := plotter.NewBarChart(group, w)
		lastBarChart.Color, _ = ParseHexColor("#ff0000") // mustAddressToColor(ap.Header.Coinbase)
		lastBarChart.LineStyle.Width = 0
		lastBarChart.XMin = float64(w) * float64(mi)
		p.Add(lastBarChart)

		if mi == 0 {
			p.Legend.Add("Miner", lastBarChart)
		}

		// collect all the unique transaction balances
		dict := map[common.Address]bool{
			ap.Header.Coinbase: true,
		}
		bals := []float64{}
		for _, tx := range ap.AppTxes {
			// dedupe
			if _, ok := dict[tx.From]; ok {
				continue
			} else {
				dict[tx.From] = true
			}
			bal, _ := lib.PrettyBalance(tx.BalanceAtParent).Float64()
			bals = append(bals, bal)

			// Add this val to the block TAB as well (order does not matter)
			tab += bal
		}

		tabSamples = append(tabSamples, tab)

		// sort the unique balances to ascending order
		sort.Float64s(bals)

		// iterate backwards
		for i := len(bals) - 1; i >= 0; i-- {
			bal := bals[i]
			vals := plotter.Values{bal}
			b, _ := plotter.NewBarChart(vals, w)
			b.Color, _ = ParseHexColor("#0000ff")
			b.LineStyle.Width = 0
			b.StackOn(lastBarChart)
			p.Add(b)
			lastBarChart = b

			if mi == 0 && i == 0 {
				p.Legend.Add("Public", b)
			}
		}
	}

	// Naive (overall, sampled set) stats about TAB values
	tabConstMean, _ := stats.Mean(tabSamples)
	constMeanLine, _ := plotter.NewLine(plotter.XYs{
		{X: 0, Y: tabConstMean},
		{X: float64(sampleSize) * float64(w), Y: tabConstMean},
	})
	// constMeanLine.LineStyle.Dashes = []vg.Length{2}
	constMeanLine.Color = colornames.Cyan

	p.Add(constMeanLine)
	p.Legend.Add("Samples Mean", constMeanLine)

	tabConstMed, _ := stats.Median(tabSamples)
	constMedLine, _ := plotter.NewLine(plotter.XYs{
		{X: 0, Y: tabConstMed},
		{X: float64(sampleSize) * float64(w), Y: tabConstMed},
	})
	// constMedLine.LineStyle.Dashes = []vg.Length{3}
	constMedLine.Color = colornames.Lightgreen
	p.Add(constMedLine)
	p.Legend.Add("Samples Median", constMedLine)

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
