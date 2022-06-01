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
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/montanaflynn/stats"
	"github.com/whilei/go-tabs-scraper/lib"
	"golang.org/x/image/colornames"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func main() {
	datadir := flag.String("datadir", "data", "Root data directory. Will be created if not existing.")
	maxSamples := flag.Int64("max-samples", 4*60*24, "Max number of samples. (default=50)")
	flag.Parse()

	matches, err := filepath.Glob(filepath.Join(*datadir, "block_*"))
	if err != nil {
		log.Fatalln(err)
	}

	sampleSize := int64(len(matches))
	if *maxSamples < int64(len(matches)) {
		sampleSize = *maxSamples
	}

	sampleWidth := float64(1)
	width := vg.Length(float64(sampleSize) * sampleWidth)
	w := width / vg.Length(sampleSize)

	estBlockTime := time.Second * 13
	estSecondsSpannedBySample := estBlockTime * time.Duration(sampleSize)

	p, _ := plot.New()
	p.Title.Text = fmt.Sprintf("Data=%s: Sampled Blocks: Miner (private) vs. Tx (public) balances, samples=%d (~%v)",
		*datadir,
		sampleSize, estSecondsSpannedBySample.Truncate(time.Second))
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

		// palette, _ := colorful.WarmPalette(len(bals))

		balsSum, _ := stats.Sum(bals)

		// iterate backwards
		for i := len(bals) - 1; i >= 0; i-- {
			bal := bals[i]
			vals := plotter.Values{bal}
			b, _ := plotter.NewBarChart(vals, w)
			// b.Color = colornames.Lightblue // ParseHexColor("#0000ff") // == blue
			// cr, cg, cb, _ := colornames.Lightblue.RGBA()

			maxBlue, baseBlue := 0.95, 0.3
			// delta := (maxBlue - baseBlue) / float64(len(bals))
			delta := (maxBlue - baseBlue) * (bal / balsSum)
			b.Color = colorful.FastLinearRgb(0.1+(delta/2), 0.1+(delta/2), baseBlue+delta)

			// b.Color = palette[i]
			b.LineStyle.Width = 0

			if mi == 0 && i == 0 {
				p.Legend.Add("Public", b)
			}

			b.StackOn(lastBarChart)
			p.Add(b)
			lastBarChart = b
		}
	}

	// Naive (overall, sampled set) stats about TAB values

	yAxisTickVals := []float64{p.Y.Min, p.Y.Max}

	tabConstMean, _ := stats.Mean(tabSamples)
	constMeanLine, _ := plotter.NewLine(plotter.XYs{
		{X: 0, Y: tabConstMean},
		{X: float64(sampleSize) * float64(w), Y: tabConstMean},
	})
	// constMeanLine.LineStyle.Dashes = []vg.Length{2}
	constMeanLine.Color = colornames.Blue

	p.Add(constMeanLine)
	p.Legend.Add(fmt.Sprintf("Samples Mean=%0.f", tabConstMean), constMeanLine)
	yAxisTickVals = append(yAxisTickVals, tabConstMean)

	tabConstMed, _ := stats.Median(tabSamples)
	constMedLine, _ := plotter.NewLine(plotter.XYs{
		{X: p.X.Min, Y: tabConstMed},
		{X: float64(sampleSize) * float64(w), Y: tabConstMed},
	})
	// constMedLine.LineStyle.Dashes = []vg.Length{3}
	constMedLine.Color = colornames.Lightgreen
	p.Add(constMedLine)
	p.Legend.Add(fmt.Sprintf("Samples Median=%0.f", tabConstMed), constMedLine)
	yAxisTickVals = append(yAxisTickVals, tabConstMed)

	// TAB percentile values

	for _, x := range []float64{1, 5, 25, 75, 95, 99} {
		v, _ := stats.Percentile(tabSamples, x)
		line, _ := plotter.NewLine(plotter.XYs{
			{X: p.X.Min, Y: v},
			{X: p.X.Max, Y: v},
		})
		yAxisTickVals = append(yAxisTickVals, v)

		// line.LineStyle.Dashes = []vg.Length{3}
		line.Width = 1
		line.Color = colornames.Black

		p.Legend.Add(fmt.Sprintf("%2.fth percentile=%0.f", x, v), line)
		p.Add(line)
	}

	// Running values (eg expected TABS)

	tabs := tabConstMed // starting value
	tabsVals := plotter.XYs{}
	for ti, tab := range tabSamples {
		if tab > tabs {
			tabs = tabs * 4097 / 4096
		} else if tab < tabs {
			tabs = tabs * 4095 / 4096
		} else {
			tabs = tabs * 4096 / 4096 // == 1 == constant
		}
		tabsVals = append(tabsVals, plotter.XY{
			X: float64(ti) * float64(w),
			Y: tabs,
		})
	}
	runningTABSLine, _ := plotter.NewLine(tabsVals)
	runningTABSLine.LineStyle.Dashes = []vg.Length{3}
	runningTABSLine.Color = colornames.Black
	p.Legend.Add(fmt.Sprintf("Simulated TABS (x=%d,y=%0.f;x=%d,y=%0.f)",
		0, tabsVals[0].Y,
		sampleSize, tabsVals[len(tabsVals)-1].Y), runningTABSLine)
	p.Add(runningTABSLine)

	// Truncate the Y limit to p95 tab
	// For ETC this makes things more visible... sadly.
	// Can comment for ETH, eg. the comment here:
	if strings.Contains(strings.ToLower(*datadir), "etc") {
		p.Y.Max, _ = stats.Percentile(tabSamples, 95)
	}

	p.Y.Tick.Marker = myTicker{vals: yAxisTickVals}

	os.MkdirAll("out", os.ModePerm)
	if err := p.Save(1200, 600, fmt.Sprintf("out/%s_stacked_balances.png",
		strings.ReplaceAll(*datadir, string(filepath.Separator), ""))); err != nil {
		log.Panic(err)
	}
}

type myTicker struct {
	vals []float64
}

func (t myTicker) Ticks(min, max float64) []plot.Tick {
	out := []plot.Tick{}
	for _, v := range t.vals {
		if v < min || v > max {
			continue
		}
		out = append(out, plot.Tick{
			Value: v,
			Label: fmt.Sprintf("%0.f", v),
		})
	}
	return out
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
