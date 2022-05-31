package main

import (
	"testing"

	"golang.org/x/exp/shiny/materialdesign/colornames"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg/draw"
)

func TestPlotInverseProportion(t *testing.T) {

	p, _ := plot.New()
	p.Title.Text = "Capital/Difficulty Equivalence Proportions (y=1/x)"
	p.Legend.Top = true

	p.X.Label.Text = "Difficulty"
	p.X.Min = 0
	p.X.Max = 5

	p.Y.Label.Text = "Capital"
	p.Y.Min = 0
	p.Y.Max = p.X.Max

	fn := plotter.NewFunction(func(f float64) float64 {
		/*
			xy = axby
			1 = ab
			1/a = b
		*/

		return float64(1) / f
	})
	fn.Samples = 1000
	fn.XMin = p.X.Min
	fn.XMax = p.X.Max

	p.Add(fn)
	p.Legend.Add("y=1/x (== xy=1)", fn)

	scatter, _ := plotter.NewScatter(plotter.XYs{plotter.XY{X: 1, Y: 1}})
	scatter.Color = colornames.Red300
	scatter.Shape = draw.CircleGlyph{}
	scatter.Radius = 5

	p.Add(scatter)
	p.Legend.Add("x=1,y=1", scatter)

	plotting, _ := plotter.NewLine(plotter.XYs{
		plotter.XY{X: p.X.Min, Y: 1},
		plotter.XY{X: p.X.Max, Y: 1},
	})
	plotting.Color = colornames.Red300

	p.Add(plotting)
	p.Legend.Add("y=1", plotting)

	scatter, _ = plotter.NewScatter(plotter.XYs{
		plotter.XY{X: 2, Y: 0.5},
		plotter.XY{X: 0.5, Y: 2},
	})
	scatter.Color = colornames.Blue300
	scatter.Shape = draw.CircleGlyph{}
	scatter.Radius = 5

	p.Add(scatter)
	p.Legend.Add("examples where y=2, x=2", scatter)

	if err := p.Save(800, 600, "1_1equivalence.png"); err != nil {
		t.Fatal(err)
	}

}
