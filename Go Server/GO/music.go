package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dhowden/tag"
	"github.com/hajimehoshi/go-mp3"
)

const SAMPLESIZE = 4

// Song estructura de una canción.
type Song struct {
	path, title, artist, album, genre string
	year                              int
	length                            time.Duration
}

// Playlist Estructura de una lista de reproducción (slice de canciones).
type Playlist struct {
	name  string
	songs []*Song
}

// Playlists Slice de Playlist.
type Playlists []*Playlist

// Variables de uso global.
var (
	WORKINGDIRECTORY, _ = os.Getwd()
	SONGSPATH           = filepath.Join(WORKINGDIRECTORY, "mp3_files")
	SUPERPLAYLIST       = Playlist{name: "SUPERPLAYLIST", songs: []*Song{}}
	TESTGENPLAYLISTS    = Playlists{}
)

// SearchPlaylist recibe el nombre de una playlist y retorna su dirección, posición en la lista y un error si se presentara.
func (pls *Playlists) SearchPlaylist(playName string) (*Playlist, int, error) {
	for i, p := range *pls {
		if p.name == playName {
			return p, i, nil
		}
	}
	return nil, -1, errors.New("playlist not found")
}

// SearchSong busca una canción en la lista de reproducción y retorna su dirección, posición en la lista y un error si se presentara.
func (pl *Playlist) SearchSong(songTitle string) (*Song, int, error) {
	for i, song := range pl.songs {
		if song.title == songTitle {
			return song, i, nil
		}
	}
	return nil, -1, errors.New("song not found A")
}

// FullSearch combina las búsquedas de SearchPlaylist y SearchSong
func (pls *Playlists) FullSearch(playName, songTitle string) (*Playlist, *Song, int, int, error) {
	playlist, i, err := pls.SearchPlaylist(playName)
	if err != nil {
		return nil, nil, -1, -1, err
	}
	song, j, err := playlist.SearchSong(songTitle)
	if err != nil {
		return playlist, nil, i, -1, err
	}
	return playlist, song, i, j, nil
}

// DeletePlaylist recibe el nombre de una playlist, la elimina y retorna el slice de playlists modificado.
func (pls *Playlists) DeletePlaylist(playName string) (*Playlists, error) {
	_, i, err := pls.SearchPlaylist(playName)
	if err != nil {
		return nil, err
	}
	*pls = append((*pls)[:i], (*pls)[i+1:]...)
	return pls, nil
}

// AddPlaylist verifica si la playlist existe, si no, la añade y retorna el slice de playlists modificado.
func (pls *Playlists) AddPlaylist(playName string) (*Playlists, error) {
	_, _, err := pls.SearchPlaylist(playName)
	if err != nil {
		*pls = append(*pls, &Playlist{playName, []*Song{}})
		return pls, nil
	}
	return nil, err
}

// DeleteSong Elimina una canción de la lista de reproducción.
func (pls *Playlists) DeleteSong(playName, songTitle string) (*Playlist, error) {
	pl, _, _, j, err := pls.FullSearch(playName, songTitle)
	if err != nil {
		return pl, err
	}
	(*pl).songs = append((*pl).songs[:j], (*pl).songs[j+1:]...)
	return pl, nil
}

// AddSong Añade una canción a la lista de reproducción.
func (pls *Playlists) AddSong(playName, songTitle string) (*Playlist, error) {
	song, _, err := SUPERPLAYLIST.SearchSong(songTitle)
	if err != nil {
		return nil, err
	}
	playlist, _, i, j, _ := pls.FullSearch(playName, song.title)
	if (i != -1) && (j == -1) {
		playlist.songs = append(playlist.songs, song)
		return playlist, nil

	}
	return nil, err
}

// FilterByYear Busca la lista de reproducción correspondiente en el slice de listas de reproducción y filtra las
// canciones de esa lista que tienen el año especificado.
func (pls *Playlists) FilterByYear(playName string, year int) ([]*Song, error) {
	pl, _, err := pls.SearchPlaylist(playName)
	if err != nil {
		return nil, err
	}

	filteredSongs := make([]*Song, 0)

	for _, song := range pl.songs {
		if song.year == year {
			filteredSongs = append(filteredSongs, song)
		}
	}

	return filteredSongs, nil
}

// FilterByDuration Busca la lista de reproducción correspondiente en el slice de listas de reproducción y filtra las
// canciones de esa lista que tienen una duración menor que la duración máxima especificada.
func (pls *Playlists) FilterByDuration(playName string, maxDuration int) ([]*Song, error) {
	pl, _, err := pls.SearchPlaylist(playName)
	if err != nil {
		return nil, err
	}

	filteredSongs := make([]*Song, 0)

	for _, song := range pl.songs {
		if song.length < time.Duration(maxDuration)*time.Minute {
			filteredSongs = append(filteredSongs, song)
		}
	}

	return filteredSongs, nil
}

// FilterByAlbum Busca la lista de reproducción correspondiente en el slice de listas de reproducción y filtra las
// canciones de esa lista que pertenecen al álbum especificado.
func (pls *Playlists) FilterByAlbum(playName string, albumName string) ([]*Song, error) {
	pl, _, err := pls.SearchPlaylist(playName)
	if err != nil {
		return nil, err
	}

	filteredSongs := make([]*Song, 0)

	for _, song := range pl.songs {
		if song.album == albumName {
			filteredSongs = append(filteredSongs, song)
		}
	}

	return filteredSongs, nil
}

func (pls *Playlists) SendMP3(conn net.Conn, playName, songTitle string) {

	_, song, _, _, err := pls.FullSearch(playName, songTitle)
	if err != nil || song == nil {
		fmt.Println("Song not found:", err)
		return
	}

	mp3File, err := os.Open(song.path)
	if err != nil {
		fmt.Println("Error opening .mp3 file:", err)
		return
	}
	defer mp3File.Close()

	// Envía el archivo MP3
	_, err = io.Copy(conn, mp3File)
	if err != nil {
		fmt.Println("Error sending .mp3 file:", err)
		return
	}

	// Envía el mensaje "EOF" para indicar el fin de la transmisión del archivo
	_, err = conn.Write([]byte("EOF"))
	if err != nil {
		fmt.Println("Error sending EOF message:", err)
		return
	}

	fmt.Printf(".mp3 file sent to %s\n", conn.RemoteAddr())
}

// SendFullPlaylistsData Itera a través de las listas de reproducción en el slice de listas de reproducción,
// y luego envía estos datos al cliente a través de la conexión.
func (pls *Playlists) SendFullPlaylistsData(conn net.Conn) {
	for _, playlist := range *pls {
		// Formatear los datos de la lista de reproducción en el formato deseado.
		playlistData := fmt.Sprintf("%s %d\n", playlist.name, len(playlist.songs))

		// Enviar los datos de la lista de reproducción al cliente.
		_, err := conn.Write([]byte(playlistData))
		if err != nil {
			fmt.Println("Error al enviar datos de la lista de reproducción:", err)
			return
		}
	}
}

// SendSongData Itera a través de las canciones en el slice y luego envía estos datos al cliente a través de la conexión.
func SendSongData(conn net.Conn, songs []*Song) {
	for _, song := range songs {
		// Formatear los datos de la canción en el formato deseado
		songData := fmt.Sprintf("%s %s %s %s \n", song.title, song.artist, song.album, song.length.String())

		// Enviar los datos de la canción al cliente
		_, err := conn.Write([]byte(songData))
		if err != nil {
			fmt.Println("Error al enviar datos de la canción:", err)
			return
		}
	}
}

// GetSongPath Devuelve el path de una canción en el directorio "mp3_files".
func GetSongPath(songTitle string) string {
	return strings.Replace(filepath.Join(SONGSPATH, songTitle+".mp3"), "\\", "/", -1)
}

// ProcessMP3File procesa un archivo MP3 dado y extrae sus metadatos, incluyendo la duración en segundos.
// Devuelve una estructura *Song que contiene los metadatos y un posible error si ocurre alguno.
func ProcessMP3File(path string) (*Song, error) {
	var song Song

	// Open the MP3 file for reading.
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
		}
	}(file)

	// Read the MP3 file metadata
	metadata, err := tag.ReadFrom(file)
	if err != nil {
		return nil, err
	}

	// Fill the song data with metadata
	song.title = metadata.Title()
	song.path = GetSongPath(metadata.Title())
	song.artist = metadata.Artist()
	song.album = metadata.Album()
	song.genre = metadata.Genre()
	song.year = metadata.Year()

	// Create a new MP3 decoder
	decoder, err := mp3.NewDecoder(file)
	if err != nil {
		return nil, err
	}

	// Calculate the duration of an MP3 file in seconds
	samples := decoder.Length() / SAMPLESIZE
	song.length = time.Duration(samples/int64(decoder.SampleRate())) * time.Second

	return &song, nil
}

// GetMP3Data recorre un directorio especificado en busca de archivos MP3 y procesa
// cada archivo para obtener sus metadatos. Luego, agrega los metadatos de cada
// canción a una lista de reproducción (Playlist) proporcionada.
func GetMP3Data(dir string, playlist *Playlist) (*Playlist, error) {
	// Recorrer el directorio para encontrar todos los archivos MP3
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Verificar si el archivo es un MP3
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".mp3") {
			// Process the MP3 file and get its metadata
			song, err := ProcessMP3File(path)
			if err != nil {
				// Handle or log the error if processing the file fails.
				fmt.Printf("Error processing file %s: %v\n", path, err)
			} else {
				// Append the processed song to the list
				playlist.songs = append(playlist.songs, song)
			}
		}
		return nil
	})

	if err != nil {
		return nil, errors.New("")
	}

	return playlist, nil

}

/*func main() {
	_, err := GetMP3Data(SONGSPATH, &SUPERPLAYLIST)
	if err != nil {
		return
	}
	fmt.Println(SUPERPLAYLIST.songs[0].title)

	// Add a playlist
	_, err = TESTGENPLAYLISTS.AddPlaylist("MyPlaylist")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("TEST 1", TESTGENPLAYLISTS[0].songs)

	// Add a song to MyPlaylist
	_, err = TESTGENPLAYLISTS.AddSong("MyPlaylist", "In the End")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("TEST 2", TESTGENPLAYLISTS[0].songs[0].title)

	// Search for a playlist
	playlist, _, err := TESTGENPLAYLISTS.SearchPlaylist("MyPlaylist")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("TEST 3: Found playlist:", playlist.name)
	}

	sonsss, _ := TESTGENPLAYLISTS.FilterByDuration("MyPlaylist", 4)
	for _, snm := range sonsss {
		fmt.Println("TEST 4", snm.path)
	}

	// Search for a song in a playlist
	song, _, err := playlist.SearchSong("In the End")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Found song:", song.title)
	}

	// Delete a song from a playlist
	_, err = TESTGENPLAYLISTS.DeleteSong("MyPlaylist", "In the End")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Deleted song from MyPlaylist")
	}
	fmt.Println(TESTGENPLAYLISTS[0].songs)

	// Delete a playlist
	_, err = TESTGENPLAYLISTS.DeletePlaylist("MyPlaylist")
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Deleted MyPlaylist")
	}

}*/
