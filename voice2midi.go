package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mjibson/go-dsp/wav"
	"github.com/urfave/cli"
)

func fft(w *wav.Wav) {
	log.Print("aaaa")
}

func readWav(path string) (*wav.Wav, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	w, err := wav.New(f)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func main() {
	app := cli.NewApp()

	// Parse command args
	app.Name = "voice2midi"
	app.Usage = "input.wav output.midi"
	app.Version = "0.0.1"
	app.Action = func(c *cli.Context) error {
		args := c.Args()

		w, err := readWav(args.Get(0))
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("sample raete: %d duration: %d samples: %d", w.SampleRate, w.Duration, w.Samples)

		fmt.Printf("sample raete: %d\nduration: %d\nsamples: %d\n", w.SampleRate, w.Duration, w.Samples)

		fft(w)
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
