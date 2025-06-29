package color

import (
	"fmt"
	"math"
	"regexp"
)

type Color struct {
	Red   int `json:"red"`
	Green int `json:"green"`
	Blue  int `json:"blue"`
}

type InterpolatedColor struct {
	From    Color `json:"from"`
	To      Color `json:"to"`
	Percent int   `json:"percent"`
}

func (c *Color) ToHex() string {
	return fmt.Sprintf("#%02X%02X%02X", c.Red, c.Green, c.Blue)
}

func FromHex(hex string) (*Color, error) {
	// regex to replace all non-hex characters
	hexRegexed := regexp.MustCompile("[^0-9a-fA-F]").ReplaceAllString(hex, "")

	lenHex := len(hexRegexed)
	if lenHex == 0 {
		return nil, fmt.Errorf("empty hex color string")
	}

	if lenHex != 6 && lenHex != 3 {
		return nil, fmt.Errorf("invalid hex color: %s", hex)
	}

	c := &Color{}

	switch lenHex {
	case 3:
		c.Red = hexCharToInt(hexRegexed[0])*16 + hexCharToInt(hexRegexed[0])
		c.Green = hexCharToInt(hexRegexed[1])*16 + hexCharToInt(hexRegexed[1])
		c.Blue = hexCharToInt(hexRegexed[2])*16 + hexCharToInt(hexRegexed[2])
	case 6:
		c.Red = hexCharToInt(hexRegexed[0])*16 + hexCharToInt(hexRegexed[1])
		c.Green = hexCharToInt(hexRegexed[2])*16 + hexCharToInt(hexRegexed[3])
		c.Blue = hexCharToInt(hexRegexed[4])*16 + hexCharToInt(hexRegexed[5])
	default:
		return nil, fmt.Errorf("invalid hex color length: %d", lenHex)
	}

	return c, nil

}

func (c *Color) ToRGB() (int, int, int) {
	return c.Red, c.Green, c.Blue
}

func (c *Color) ToRGBString() string {
	return fmt.Sprintf("rgb(%d, %d, %d)", c.Red, c.Green, c.Blue)
}

func (c *Color) Interpolate(to Color, percent int) Color {
	if percent < 0 || percent > 100 {
		return Color{} // Return zero value if percent is out of bounds
	}

	ratio := float64(percent) / 100.0

	red := c.Red + int(math.Abs(float64(to.Red-c.Red))*ratio)
	green := c.Green + int(math.Abs(float64(to.Green-c.Green))*ratio)
	blue := c.Blue + int(math.Abs(float64(to.Blue-c.Blue))*ratio)

	return Color{Red: red, Green: green, Blue: blue}
}

func (c *Color) String() string {
	return fmt.Sprintf("HexColor(R: %d, G: %d, B: %d)", c.Red, c.Green, c.Blue)
}

func NewColor(red, green, blue int) Color {
	return Color{
		Red:   red,
		Green: green,
		Blue:  blue,
	}
}

func hexCharToInt(c byte) int {
	if c >= '0' && c <= '9' {
		return int(c - '0')
	}
	if c >= 'a' && c <= 'f' {
		return int(c - 'a' + 10)
	}
	if c >= 'A' && c <= 'F' {
		return int(c - 'A' + 10)
	}
	return 0 // Error case
}
