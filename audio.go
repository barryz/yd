package main

import (
	"net/http"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

// Audio word audio abstracion.
type Audio interface {
	Play(url string, done chan struct{}) error
}

// USAudio us speech audio.
type USAudio struct {
	word string
	done chan struct{}
}

// NewUSAudio creates an new us speech audio.
func NewUSAudio(word string) Audio {
	return &USAudio{
		word: word,
		done: make(chan struct{}),
	}
}

// Play play audio.
func (ua *USAudio) Play(url string, done chan struct{}) error {
	defer func() {
		recover()
	}()

	defer func() { done <- struct{}{} }()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	streamer, format, err := mp3.Decode(resp.Body)
	if err != nil {
		return err
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		ua.done <- struct{}{}
	})))

	// waiting for finish audio play
	<-time.After(300 * time.Millisecond)
	<-ua.done
	return nil
}
