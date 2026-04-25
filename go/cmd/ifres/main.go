// Package main implements ifres, a command-line indicator frequency response calculator.
//
// It reads a JSON settings file containing indicator definitions,
// creates indicator instances, determines each indicator's warmup period,
// and calculates the frequency response with signal length 1024.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"zpano/indicators/core"
	"zpano/indicators/core/frequencyresponse"
	"zpano/indicators/factory"
)

const (
	signalLength                = 1024
	maxWarmup                   = 10000
	phaseDegreesUnwrappingLimit = 179.0
)

// settingsEntry represents a single indicator entry in the settings JSON file.
type settingsEntry struct {
	Identifier core.Identifier `json:"identifier"`
	Params     json.RawMessage `json:"params"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: ifres <settings.json>\n")
		os.Exit(1)
	}

	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading settings file: %v\n", err)
		os.Exit(1)
	}

	var entries []settingsEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		fmt.Fprintf(os.Stderr, "error parsing settings file: %v\n", err)
		os.Exit(1)
	}

	for _, e := range entries {
		params := string(e.Params)

		// Create a probe instance to determine warmup period.
		probe, err := factory.New(e.Identifier, params)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating indicator %s: %v\n", e.Identifier, err)
			os.Exit(1)
		}

		probeUpdater, ok := probe.(frequencyresponse.Updater)
		if !ok {
			fmt.Fprintf(os.Stderr, "indicator %s does not satisfy frequencyresponse.Updater\n", e.Identifier)
			os.Exit(1)
		}

		warmup := detectWarmup(probeUpdater, probe)

		// Create a fresh instance for the actual calculation.
		ind, err := factory.New(e.Identifier, params)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating indicator %s: %v\n", e.Identifier, err)
			os.Exit(1)
		}

		updater, ok := ind.(frequencyresponse.Updater)
		if !ok {
			fmt.Fprintf(os.Stderr, "indicator %s does not satisfy frequencyresponse.Updater\n", e.Identifier)
			os.Exit(1)
		}

		fr, err := frequencyresponse.Calculate(signalLength, updater, warmup, phaseDegreesUnwrappingLimit)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error calculating frequency response for %s: %v\n", e.Identifier, err)
			os.Exit(1)
		}

		printFrequencyResponse(fr, warmup)
	}
}

// detectWarmup feeds zeros into the indicator until it is primed,
// returning the number of samples needed.
func detectWarmup(updater frequencyresponse.Updater, ind core.Indicator) int {
	for i := range maxWarmup {
		if ind.IsPrimed() {
			return i
		}

		updater.Update(0)
	}

	return maxWarmup
}

// printFrequencyResponse prints a summary of the frequency response data.
func printFrequencyResponse(fr *frequencyresponse.FrequencyResponse, warmup int) {
	fmt.Printf("=== %s (warmup=%d) ===\n", fr.Label, warmup)
	fmt.Printf("  Spectrum length: %d\n", len(fr.NormalizedFrequency))

	printComponent("PowerPercent", fr.PowerPercent)
	printComponent("PowerDecibel", fr.PowerDecibel)
	printComponent("AmplitudePercent", fr.AmplitudePercent)
	printComponent("AmplitudeDecibel", fr.AmplitudeDecibel)
	printComponent("PhaseDegrees", fr.PhaseDegrees)
	printComponent("PhaseDegreesUnwrapped", fr.PhaseDegreesUnwrapped)

	fmt.Println()
}

// printComponent prints min, max, and a few sample values from a frequency response component.
func printComponent(name string, c frequencyresponse.Component) {
	fmt.Printf("  %-25s min=%10.4f  max=%10.4f", name, c.Min, c.Max)

	n := len(c.Data)
	if n == 0 {
		fmt.Println()
		return
	}

	// Print first 3 and last 3 values as a preview.
	preview := 3
	if n <= preview*2 {
		fmt.Printf("  data=%v", c.Data)
	} else {
		fmt.Printf("  data=[%.4f %.4f %.4f ... %.4f %.4f %.4f]",
			c.Data[0], c.Data[1], c.Data[2],
			c.Data[n-3], c.Data[n-2], c.Data[n-1])
	}

	fmt.Println()
}
