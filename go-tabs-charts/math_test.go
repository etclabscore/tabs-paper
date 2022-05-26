package main

import (
	"fmt"
	"math"
	"testing"

	"golang.org/x/exp/shiny/materialdesign/colornames"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
)

func TestDifficultyChangeRateRange(t *testing.T) {
	p, _ := plot.New()
	p.Title.Text = "Difficulty Rise/Fall Potential"

	p.X.Label.Text = "Time"
	p.X.Min = 0
	p.X.Max = (60 / 14) * 60

	p.Y.Label.Text = "Difficulty"
	p.Y.Min = 0
	p.Y.Max = 2

	for i := float64(2); i >= -42; i-- {
		i := i
		fn := plotter.NewFunction(func(f float64) float64 {
			num := 2048 + i
			rate := num / 2048
			return math.Pow(rate, f/14)
		})
		if i > 0 {
			fn.Color = colornames.Green200
		} else if i == 0 {
			fn.Color = colornames.Grey200
		} else {
			// < 0
			fn.Color = colornames.Red200
		}

		p.Legend.Add(fmt.Sprintf("%0.f", i), fn)
		p.Add(fn)
	}

	if err := p.Save(800, 600, "difficulty_change.png"); err != nil {
		t.Fatal(err)
	}
}

func TestTABSChangeRateRange(t *testing.T) {
	p, _ := plot.New()
	p.Title.Text = "TABS Rise/Fall Potential"

	p.X.Label.Text = "Time"
	p.X.Min = 0
	p.X.Max = (60 / 14) * 60

	p.Y.Label.Text = "Difficulty"
	p.Y.Min = 0
	p.Y.Max = 2

	for i := float64(1); i >= -1; i-- {
		i := i
		fn := plotter.NewFunction(func(f float64) float64 {
			num := 4096 + i
			rate := num / 4096
			return math.Pow(rate, f)
		})
		if i > 0 {
			fn.Color = colornames.Green200
		} else if i == 0 {
			fn.Color = colornames.Grey200
		} else {
			// < 0
			fn.Color = colornames.Red200
		}

		p.Legend.Add(fmt.Sprintf("%0.f", i), fn)
		p.Add(fn)
	}

	if err := p.Save(800, 600, "tabs_change.png"); err != nil {
		t.Fatal(err)
	}
}
