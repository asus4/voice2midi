package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gomidi/midi/smf"
	"github.com/gomidi/midi/smf/smfwriter"
	"github.com/mjibson/go-dsp/wav"

	"github.com/urfave/cli"
)

func convert(wavFile string, midiFile string) error {
	// Read file
	file, err := os.Open(wavFile)
	if err != nil {
		return err
	}
	defer file.Close()

	w, err := wav.New(file)
	if err != nil {
		return err
	}

	// Print into
	fmt.Println("Converting")
	fmt.Printf("sample rate: %d\n", w.SampleRate)
	fmt.Printf("duration: %d\n", w.Duration)
	fmt.Printf("samples: %d\n", w.Samples)
	fmt.Printf("channels: %d\n", w.Header.NumChannels)

	spectrums, err := makeSpectrums(w, 1024)
	if err != nil {
		return err
	}

	// log.Print(spectrums[len(spectrums)/2])

	converter := voiceConverter{
		sampleRate: w.SampleRate,
		bufferSize: 1024,
		tpq:        smf.MetricTicks(960), // Use default tick: 960
		peaks:      findPeaks(spectrums, 32),
		maxPeak:    maxVolume(spectrums) / 2,
		channel:    1,
	}

	log.Printf("maxpeak: %f", converter.maxPeak)

	err = smfwriter.WriteFile(midiFile, converter.write, smfwriter.NumTracks(1), smfwriter.TimeFormat(converter.tpq))
	if err != nil {
		return err
	}

	return nil
}

func main() {
	// command args
	app := cli.NewApp()
	app.Name = "voice2midi"
	app.Usage = "input.wav output.midi"
	app.Version = "0.0.1"
	app.Action = func(c *cli.Context) error {
		args := c.Args()
		return convert(args.Get(0), args.Get(1))
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
