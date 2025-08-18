package cmd

import (
	"fmt"
	"math"
	"strconv"

	"github.com/spf13/cobra"
)

var SqrtCmd = &cobra.Command{
	Use:   "sqrt <number>",
	Short: "Calculate the square root of a number",
	RunE:  runSqrtCmd,
	Args:  cobra.ExactArgs(1),
}

func init() {
	SqrtCmd.Flags().BoolP("quake-3", "Q", false, "Use the magic number Quake Three method")
	SqrtCmd.Flags().BoolP("heron", "H", false, "Use Heron's method")

}

func runSqrtCmd(cmd *cobra.Command, args []string) error {
	num64, err := strconv.ParseFloat(args[0], 32)
	if err != nil {
		return err
	}

	num := float64(num64)

	if num < 0 {
		return fmt.Errorf("cannot calculate square root of a negative number")
	}

	quakeThree, _ := cmd.Flags().GetBool("quake-3")
	useHeronsMethod, _ := cmd.Flags().GetBool("heron")

	var result any

	if quakeThree {
		result = quakeThreeMethod(float32(num))
	} else if useHeronsMethod {
		result = heronsMethod(num)
	} else {
		result = float32(math.Sqrt(float64(num)))
	}

	fmt.Printf("%f\n", result)

	return nil
}

func quakeThreeMethod(num float32) float32 {

	half := float32(num * 0.5)
	u := struct {
		N float32
		I int32
	}{N: 0, I: 0}

	u.N = num
	u.I = 0x5f375a86 - (u.I >> 1)
	u.N = u.N * (1.5 - half*u.N*u.N)
	u.N = u.N * (1.5 - half*u.N*u.N)
	u.N = u.N * (1.5 - half*u.N*u.N)

	return 1 / (num * u.N)
}

func heronsMethod(s float64) float64 {
	firstGuess := s / 2
	currentStep := firstGuess
	epsilon := s * 1.0e-9

	for {
		nextStep := (1.0 / 2.0) * (currentStep + (s / currentStep))
		if math.Abs(nextStep-currentStep) < epsilon {
			break
		}
		currentStep = nextStep
	}
	return currentStep
}
