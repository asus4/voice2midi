package main

import (
	"log"
	"math"
	"time"

	"github.com/gomidi/midi/midimessage/channel"
	"github.com/gomidi/midi/midimessage/meta"
	"github.com/gomidi/midi/smf"
)

type voiceConverter struct {
	sampleRate uint32
	bufferSize uint16
	tpq        smf.MetricTicks // ticks per quarter
	peaks      [](map[float64]int)
	maxPeak    float64
	channel    channel.Channel
}

func freq2notef(freq float64) float64 {
	note := 12.0*math.Log2(freq/440.0) + 69.0
	return math.Max(note, 0)
}

func freq2note(freq float64) uint8 {
	note := freq2notef(freq)
	note = math.Round(note)
	note = math.Min(note, 127)
	return uint8(note)
}

// https://stackoverflow.com/questions/4364823/how-do-i-obtain-the-frequencies-of-each-value-in-an-fft
func (c voiceConverter) fft2freq(band uint32) float64 {
	return float64(band) * float64(c.sampleRate) / float64(c.bufferSize)
}

func (c voiceConverter) fft2note(band uint32, volume float64) channel.NoteOn {
	hz := c.fft2freq(band)
	return channel.Channel(c.channel).NoteOn(freq2note(hz), uint8(volume/c.maxPeak*127.0))
}

func (c voiceConverter) peaks2messages(index int) []channel.NoteOn {
	var messages []channel.NoteOn
	for volume, n := range c.peaks[index] {
		m := c.fft2note(uint32(n), volume)
		if m.Key() > 1 && m.Velocity() > 0 {
			messages = append(messages, m)
		}
	}
	return messages
}

func (c voiceConverter) sampleTime(sample uint64) time.Duration {
	t := float64(sample) / float64(c.sampleRate)
	return time.Duration(t * float64(time.Second))
}

func findSameNote(arr []channel.NoteOn, note channel.NoteOn) bool {
	for _, n := range arr {
		if n.Key() == note.Key() {
			return true
		}
	}
	return false
}

func compareNotes(current []channel.NoteOn, next []channel.NoteOn) ([]channel.NoteOn, []channel.NoteOff) {
	// curernt true && next true => remain
	// current false && next true => note on
	// current true && next false => note off
	var noteons []channel.NoteOn
	var noteoffs []channel.NoteOff

	for _, n := range next {
		if !findSameNote(current, n) {
			noteons = append(noteons, n)
		}
	}
	for _, c := range current {
		if !findSameNote(next, c) {
			ch := channel.Channel(c.Channel())
			noteoffs = append(noteoffs, ch.NoteOff(c.Key()))
		}
	}
	return noteons, noteoffs
}

func (c voiceConverter) write(wr smf.Writer) {
	const TEMPO = 120
	var messages []channel.NoteOn
	tick := uint32(0)

	wr.SetDelta(0)
	wr.Write(meta.Tempo(TEMPO))
	wr.SetDelta(1)

	// add each channels
	for i := range c.peaks {
		next := c.peaks2messages(i)
		noteOns, noteOffs := compareNotes(messages, next)

		if len(noteOns) > 0 || len(noteOffs) > 0 {
			log.Printf("------ %d", i)
			log.Print(noteOns)
			log.Print(noteOffs)
		}

		for _, m := range noteOffs {
			wr.Write(m)
		}
		for _, m := range noteOns {
			wr.Write(m)
		}
		messages = next

		// set delta
		t := c.tpq.Ticks(TEMPO, c.sampleTime(uint64(i)*uint64(c.bufferSize)))
		wr.SetDelta(t - tick)
		tick = t
	}

	// note off last
	for _, m := range messages {
		wr.Write(c.channel.NoteOff(m.Key()))
	}

	// close track
	wr.SetDelta(1)
	wr.Write(meta.EndOfTrack)
}
