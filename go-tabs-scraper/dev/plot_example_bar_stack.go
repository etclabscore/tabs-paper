package main

import (
	"image/color"
	"log"
	"os"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
)

func main() {
	os.MkdirAll("testdata", os.ModePerm)

	// Create the plot values and labels.
	values := plotter.Values{0.5, 10, 20, 30}
	verticalLabels := []string{"A", "B", "C", "D"}
	horizontalLabels := []string{"Label A", "Label B", "Label C", "Label D"}

	// Create a vertical BarChart
	p1 := plot.New()
	verticalBarChart, err := plotter.NewBarChart(values, 0.5*vg.Centimeter)
	if err != nil {
		log.Panic(err)
	}
	p1.Add(verticalBarChart)
	p1.NominalX(verticalLabels...)
	err = p1.Save(100, 100, "testdata/verticalBarChart.png")
	if err != nil {
		log.Panic(err)
	}

	// Create a horizontal BarChart
	p2 := plot.New()
	horizontalBarChart, err := plotter.NewBarChart(values, 0.5*vg.Centimeter)
	horizontalBarChart.Horizontal = true // Specify a horizontal BarChart.
	if err != nil {
		log.Panic(err)
	}
	p2.Add(horizontalBarChart)
	p2.NominalY(horizontalLabels...)
	err = p2.Save(100, 100, "testdata/horizontalBarChart.png")
	if err != nil {
		log.Panic(err)
	}

	// Now, make a different type of BarChart.
	groupA := plotter.Values{20, 35, 30, 35, 27}
	groupB := plotter.Values{25, 32, 34, 20, 25}
	groupC := plotter.Values{12, 28, 15, 21, 8}
	groupD := plotter.Values{30, 42, 6, 9, 12}

	p := plot.New()
	p.Title.Text = "Bar chart"
	p.Y.Label.Text = "Heights"

	w := vg.Points(8)

	barsA, err := plotter.NewBarChart(groupA, w)
	if err != nil {
		log.Panic(err)
	}
	barsA.Color = color.RGBA{R: 255, A: 255}
	barsA.Offset = -w / 2

	barsB, err := plotter.NewBarChart(groupB, w)
	if err != nil {
		log.Panic(err)
	}
	barsB.Color = color.RGBA{R: 196, G: 196, A: 255}
	barsB.Offset = w / 2

	barsC, err := plotter.NewBarChart(groupC, w)
	if err != nil {
		log.Panic(err)
	}
	barsC.XMin = 6
	barsC.Color = color.RGBA{B: 255, A: 255}
	barsC.Offset = -w / 2

	barsD, err := plotter.NewBarChart(groupD, w)
	if err != nil {
		log.Panic(err)
	}
	barsD.Color = color.RGBA{B: 255, R: 255, A: 255}
	barsD.XMin = 6
	barsD.Offset = w / 2

	p.Add(barsA, barsB, barsC, barsD)
	p.Legend.Add("A", barsA)
	p.Legend.Add("B", barsB)
	p.Legend.Add("C", barsC)
	p.Legend.Add("D", barsD)
	p.Legend.Top = true
	p.NominalX("Zero", "One", "Two", "Three", "Four", "",
		"Six", "Seven", "Eight", "Nine", "Ten")

	p.Add(plotter.NewGlyphBoxes())
	err = p.Save(300, 250, "testdata/barChart2.png")
	if err != nil {
		log.Panic(err)
	}

	// Now, make a stacked BarChart.
	p = plot.New()
	p.Title.Text = "Bar chart"
	p.Y.Label.Text = "Heights"

	w = vg.Points(15)

	barsA, err = plotter.NewBarChart(groupA, w)
	if err != nil {
		log.Panic(err)
	}
	barsA.Color = color.RGBA{R: 255, A: 255}
	// barsA.Offset = -w / 2

	barsB, err = plotter.NewBarChart(groupB, w)
	if err != nil {
		log.Panic(err)
	}
	barsB.Color = color.RGBA{R: 196, G: 196, A: 255}
	barsB.StackOn(barsA)

	barsC, err = plotter.NewBarChart(groupC, w)
	if err != nil {
		log.Panic(err)
	}
	// barsC.Offset = w / 2
	barsC.Color = color.RGBA{B: 255, A: 255}
	barsC.StackOn(barsB)

	barsD, err = plotter.NewBarChart(groupD, w)
	if err != nil {
		log.Panic(err)
	}
	barsD.StackOn(barsC)
	barsD.Color = color.RGBA{B: 255, R: 255, A: 255}

	p.Add(barsA, barsB, barsC, barsD)
	p.Legend.Add("A", barsA)
	p.Legend.Add("B", barsB)
	p.Legend.Add("C", barsC)
	p.Legend.Add("D", barsD)
	p.Legend.Top = true
	p.NominalX("Zero", "One", "Two", "Three", "Four")

	p.Add(plotter.NewGlyphBoxes())
	err = p.Save(250, 250, "testdata/stackedBarChart.png")
	if err != nil {
		log.Panic(err)
	}
}
