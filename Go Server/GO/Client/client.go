package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

const serverAddr = "localhost:8081"
const bufferSize = 32768

func main() {
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer conn.Close()

	// Wait for user input to enter the song name
	fmt.Print("Enter the song name: ")
	var songName string
	fmt.Scanln(&songName)

	// Send the song name as a request to the server
	_, err = conn.Write([]byte("PLAYS::" + "NULL::" + songName))
	if err != nil {
		fmt.Println("Error sending request to server:", err)
		return
	}

	// Receive the server's response
	response := make([]byte, bufferSize)
	n, err := conn.Read(response)
	if err != nil {
		if err != io.EOF {
			fmt.Println("Error reading response from server:", err)
		}
		return
	}

	// Check if the response is "Invalid Input"
	if strings.TrimSpace(string(response[:n])) == "Invalid Input" {
		fmt.Println("Server responded with 'Invalid Input'. Please check your request.")
		return
	}

	// Create a buffer to hold the received song data
	var songData []byte

	buf := make([]byte, bufferSize)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Println("Error reading from server:", err)
			}
			break
		}

		// Append the received data to the songData slice
		songData = append(songData, buf[:n]...)
	}

	// Create a channel for stopping song playback
	stopPlayback := make(chan struct{})

	// Start a Goroutine to play the received song data
	go playSong(songData, stopPlayback)

	// Wait for user input or other conditions to stop playback
	// For example, you can use fmt.Scanln() to wait for user input to stop playback
	var input string
	fmt.Println("Press Enter to stop playback...")
	fmt.Scanln(&input)

	// Send a stop signal to the playback Goroutine
	close(stopPlayback)

	// You can optionally wait for the playback Goroutine to finish
}

func playSong(songData []byte, stopPlayback chan struct{}) {
	// Decode the song data using the mp3 decoder
	streamer, format, err := mp3.Decode(io.NopCloser(bytes.NewReader(songData)))
	if err != nil {
		fmt.Println("Error decoding MP3 data:", err)
		return
	}

	// Initialize the audio player
	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		fmt.Println("Error initializing audio player:", err)
		return
	}

	// Play the decoded audio stream
	done := make(chan struct{})
	go func() {
		speaker.Play(beep.Seq(streamer, beep.Callback(func() {
			close(done)
		})))
	}()

	// Wait for either playback to finish or a stop signal
	select {
	case <-done:
		fmt.Println("Song playback completed.")
	case <-stopPlayback:
		fmt.Println("Song playback stopped.")
		speaker.Clear() // Stop playback immediately
	}
}
