package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

type Playlists []Playlist

// Variables globales de uso general.
var (
	workingDirectory, _ = os.Getwd()
	projectDir          = filepath.Dir(workingDirectory)
	songsPath           = filepath.Join(projectDir, "mp3_files")
	generalPlaylist     = Playlists{}
)

// Devuelve el path de una canción en el directorio "mp3_files".
func getSongPath(name string) string {
	return strings.Replace((filepath.Join(songsPath, (name + ".mp3"))), "\\", "/", -1)
}

// Crear una lista de reproducción.
func (ps *Playlists) addPlaylist(name string) {
	*ps = append(*ps, Playlist{name, []Song{}})
}

// Busca una lista de reproducción.
func (ps *Playlists) searchPlaylist(name string) (bool, *Playlist) {
	for _, pls := range *ps {
		if pls.name == name {
			return true, &pls
		}
	}
	return false, nil
}

func main() {
	//filePath := "C:\Users\aleji\OneDrive\Desktop\Clases\Semestre II 2023\Lenguajes\Archivos de proyecto\mp3_files\WildSide.mp3"
	filePath := getSongPath("Alone_-_Color_Out")
	println(os.Getwd())
	println(filePath)
	println(getSongPath("Alone_-_Color_Out"))

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Fatalf("File does not exist: %s", filePath)
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	streamer, format, err := mp3.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	fmt.Println("Playing audio...")
	<-done
	fmt.Println("Audio playback complete.")
}
