package gauge

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/sqshq/sampler/component"
	"github.com/sqshq/sampler/config"
	"github.com/sqshq/sampler/console"
	"github.com/sqshq/sampler/data"
	"image"
	"math"
	"strconv"
)

const (
	MinValueLabel = "min"
	MaxValueLabel = "max"
	CurValueLabel = "cur"
)

type Gauge struct {
	*component.Component
	minValue float64
	maxValue float64
	curValue float64
	color    ui.Color
	scale    int
}

func NewGauge(c config.GaugeConfig) *Gauge {

	gauge := Gauge{
		Component: component.NewComponent(c.ComponentConfig, config.TypeGauge),
		scale:     *c.Scale,
		color:     *c.Color,
	}

	go func() {
		for {
			select {
			case sample := <-gauge.SampleChannel:
				gauge.ConsumeSample(sample)
			}
		}
	}()

	return &gauge
}

func (g *Gauge) ConsumeSample(sample data.Sample) {

	float, err := strconv.ParseFloat(sample.Value, 64)
	if err != nil {
		// TODO handle in Component
	}

	switch sample.Label {
	case MinValueLabel:
		g.minValue = float
		break
	case MaxValueLabel:
		g.maxValue = float
		break
	case CurValueLabel:
		g.curValue = float
		break
	}
}

func (g *Gauge) Draw(buf *ui.Buffer) {

	g.Block.Draw(buf)

	percent := 0.0
	if g.curValue != 0 && g.maxValue != g.minValue {
		percent = (100 * g.curValue) / (g.maxValue - g.minValue)
	}

	label := fmt.Sprintf("%v%% (%v)", formatValue(percent, g.scale), g.curValue)

	// draw bar
	barWidth := int((percent / 100) * float64(g.Inner.Dx()))
	if barWidth == 0 {
		barWidth = 1
	} else if barWidth > g.Dx()-2 {
		barWidth = g.Dx() - 2
	}
	buf.Fill(
		ui.NewCell(console.SymbolVerticalBar, ui.NewStyle(g.color)),
		image.Rect(g.Inner.Min.X+1, g.Inner.Min.Y, g.Inner.Min.X+barWidth, g.Inner.Max.Y),
	)

	// draw label
	labelXCoordinate := g.Inner.Min.X + (g.Inner.Dx() / 2) - int(float64(len(label))/2)
	labelYCoordinate := g.Inner.Min.Y + ((g.Inner.Dy() - 1) / 2)
	if labelYCoordinate < g.Inner.Max.Y {
		for i, char := range label {
			style := ui.NewStyle(console.ColorWhite)
			if labelXCoordinate+i+1 <= g.Inner.Min.X+barWidth {
				style = ui.NewStyle(console.ColorWhite, ui.ColorClear)
			}
			buf.SetCell(ui.NewCell(char, style), image.Pt(labelXCoordinate+i, labelYCoordinate))
		}
	}
}

// TODO extract to utils
func formatValue(value float64, scale int) string {
	if math.Abs(value) == math.MaxFloat64 {
		return "Inf"
	} else {
		format := "%." + strconv.Itoa(scale) + "f"
		return fmt.Sprintf(format, value)
	}
}
