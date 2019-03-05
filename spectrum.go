package main

import (
	"math"
	"math/cmplx"
	"sort"

	"github.com/mjibson/go-dsp/fft"
	"github.com/mjibson/go-dsp/spectral"
	"github.com/mjibson/go-dsp/wav"
	"github.com/mjibson/go-dsp/window"
)

func mergeChannels(samples []float32, channels uint16) []float64 {
	lenth := len(samples) / int(channels)
	merged := make([]float64, lenth)

	var sum float64
	for i := 0; i < lenth; i++ {
		sum = 0.0
		for j := 0; j < int(channels); j++ {
			sum += float64(samples[i*int(channels)+j])
		}
		merged[i] = sum / float64(channels)
	}

	return merged
}

func makeSpectrums(w *wav.Wav, windowSize int) ([][]float64, error) {

	// Read float buffers
	allSamples, err := w.ReadFloats(w.Samples)
	if err != nil {
		return nil, err
	}
	singleSamples := mergeChannels(allSamples, w.Header.NumChannels)

	// Apply window function [][]float64
	windows := spectral.Segment(singleSamples, windowSize, 0)
	for _, buffer := range windows {
		window.Apply(buffer, window.Hann)
	}

	// Make spectrum
	spectrums := make([][]float64, len(windows))
	for i, window := range windows {
		// Length is N/2, due to Nyquist frequency
		spectrums[i] = make([]float64, len(window)/2)
		spectrum := fft.FFTReal(window)
		for j := 0; j < len(window)/2; j++ {
			spectrums[i][j] = cmplx.Abs(spectrum[j])
		}
	}
	return spectrums, nil
}

func findPeaks(spectrums [][]float64, maxCount int) [](map[float64]int) {
	peaks := make([]map[float64]int, len(spectrums))
	for i := 0; i < len(spectrums); i++ {
		peaks[i] = limitPeaks(findPeak(spectrums[i]), maxCount)
	}
	return peaks
}

func maxVolume(spectrums [][]float64) float64 {
	n := 0.0
	for _, spectrum := range spectrums {
		for _, s := range spectrum {
			n = math.Max(n, s)
		}
	}
	return n
}

func findPeak(spectrum []float64) map[float64]int {
	// Get all peaks
	var gains []float64
	peaks := make(map[float64]int)
	up := true
	n := 0.0

	for i := 0; i < len(spectrum); i++ {
		if n <= spectrum[i] {
			up = true
		} else {
			if up {
				peaks[n] = i - 1
				gains = append(gains, n)
			}
			up = false
		}
		n = spectrum[i]
	}
	return peaks
}

func limitPeaks(peaks map[float64]int, maxCount int) map[float64]int {
	if len(peaks) < maxCount {
		return peaks
	}

	// get gain keys
	keys := make([]float64, len(peaks))
	i := 0
	for k := range peaks {
		keys[i] = k
		i++
	}

	// limit biggest
	sort.Float64s(keys)
	keys = keys[len(keys)-maxCount-1:]

	newpeaks := make(map[float64]int)
	for _, k := range keys {
		newpeaks[k] = peaks[k]
	}
	return newpeaks
}
