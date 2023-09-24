package main

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

const (
	LISTENER = "localhost:8081"
	BUFFER   = 10000
)

type Message struct {
	from    string
	payload []byte
}

type Server struct {
	listenAddr   string
	ln           net.Listener
	quitCh       chan struct{}
	msgCh        chan Message
	genPlaylists []*Playlists
}

func NewServer(listenAddr string) *Server {
	return &Server{
		listenAddr:   listenAddr,
		quitCh:       make(chan struct{}),
		msgCh:        make(chan Message, BUFFER),
		genPlaylists: []*Playlists{},
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer func(ln net.Listener) {
		err := ln.Close()
		if err != nil {

		}
	}(ln)
	s.ln = ln

	go s.acceptLoop()

	<-s.quitCh
	close(s.msgCh)

	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			fmt.Println("accept error:", err)
			continue
		}

		fmt.Println("new connection to the server:", conn.RemoteAddr())

		clientPlaylists := &Playlists{&SUPERPLAYLIST} // Agregar a SUPERPLAYLIST en el slice de playlists del cliente para que se inicialice la app con las canciones disponibles
		s.genPlaylists = append(s.genPlaylists, clientPlaylists)

		go s.readLoop(conn, clientPlaylists)
	}
}

func SendSongList(conn net.Conn) { //Para pasarle al GUI la lista de canciones para que genere los botones de la ventana principal para cada canción disponible
	// Recorre todas las canciones en SUPERPLAYLIST y envía sus datos al cliente
	for _, song := range SUPERPLAYLIST.songs {
		// Formatear los datos de la canción para enviar al cliente
		songInfo := fmt.Sprintf("%s::%s::%d::%s\n", song.title, song.artist, song.year, song.length.String())

		// Enviar los datos de la canción al cliente
		_, err := conn.Write([]byte(songInfo))
		if err != nil {
			fmt.Println("Error al enviar datos de la canción:", err)
			return
		}
	}
}

func (s *Server) readLoop(conn net.Conn, clientPlaylists *Playlists) {

	buf := make([]byte, BUFFER)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("read error:", err)
			return // Terminar la función si ocurre un error en la lectura
		}

		// Obtener el mensaje enviado por el cliente.
		message := string(buf[:n])
		fmt.Println("Received message from client:", message)
		parts := strings.Split(message, "::")

		if message == "GET_SONG_LIST" {
			SendSongList(conn)
			continue // Salta al siguiente ciclo del bucle para esperar otra solicitud
		} else if len(parts) < 4 {
			fmt.Println(message)
			continue
		}

		code := parts[0] //(PLAY_S,ADD_S,DEL_S,ADD_P,DEL_P)
		playName := parts[1]
		songTitle := parts[2]
		extraCode := parts[3]

		// Crear una instancia con la dirección remota del cliente.
		s.msgCh <- Message{
			from:    conn.RemoteAddr().String(),
			payload: []byte(message),
		}

		// Ejecutar acción dependiendo del contenido del mensaje.
		switch code {

		// Ejecuta la búsqueda de una canción y envía sus datos.
		case "SRH_S":
			var songs = make([]*Song, 1)
			song, _, err := SUPERPLAYLIST.SearchSong(songTitle)
			if err != nil {
				fmt.Println(err)
			}
			songs = append(songs, song)
			SendSongData(conn, songs)

		// Ejecuta la búsqueda de una playlist y envía sus datos.
		case "SRH_P":
			playlist, _, err := clientPlaylists.SearchPlaylist(playName)
			if err != nil {
				fmt.Println(err)
			}
			SendSongData(conn, playlist.songs)

		// Envía los datos de todas las playlists del cliente.
		case "SHW_FP":
			clientPlaylists.SendFullPlaylistsData(conn)

		// Envía los bytes de un archivo MP3 por el servidor.
		case "PLAY_S":
			clientPlaylists.SendMP3(conn, playName, songTitle)

		// Añade una canción a una playlist.
		case "ADD_S":
			playlist, err := clientPlaylists.AddSong(playName, songTitle)
			if err != nil {
				fmt.Println(err)
			}
			SendSongData(conn, playlist.songs)
		// Elimina una canción de una playlist.
		case "DEL_S":
			playlist, err := clientPlaylists.DeleteSong(playName, songTitle)
			if err != nil {
				fmt.Println(err)
			}
			SendSongData(conn, playlist.songs)
		// Añade una nueva playlist al slice de playlists.
		case "ADD_P":
			_, err := clientPlaylists.AddPlaylist(playName)
			if err != nil {
				fmt.Println(err)
			}
			clientPlaylists.SendFullPlaylistsData(conn)
		// Elimina una playlist en el slice de playlists.
		case "DEL_P":
			_, err := clientPlaylists.DeletePlaylist(playName)
			if err != nil {
				fmt.Println(err)
			}
			clientPlaylists.SendFullPlaylistsData(conn)
		// Filtra la playlist por año.
		case "FLT_Y":
			year, _ := strconv.Atoi(extraCode)
			songs, err := clientPlaylists.FilterByYear(playName, year)
			if err != nil {
				fmt.Println(err)
			}
			SendSongData(conn, songs)
		// Filtra la playlist por duración.
		case "FLT_L":
			duration, _ := strconv.Atoi(extraCode)
			songs, err := clientPlaylists.FilterByDuration(playName, duration)
			if err != nil {
				fmt.Println(err)
			}
			SendSongData(conn, songs)
		// Filtra la playlist por album.
		case "FLT_A":
			songs, err := clientPlaylists.FilterByAlbum(playName, extraCode)
			if err != nil {
				fmt.Println(err)
			}
			SendSongData(conn, songs)
		// Caso base:
		default:
			_, err = conn.Write([]byte("Invalid Input"))
			if err != nil {
				fmt.Println(err)
			}
		}

	}
}

func main() {
	// Cargar canciones en la SUPERPLAYLIST antes de iniciar el servidor
	_, err := GetMP3Data(SONGSPATH, &SUPERPLAYLIST)
	if err != nil {
		fmt.Println("Error al obtener datos de MP3:", err)
		return
	}

	if len(SUPERPLAYLIST.songs) == 0 {
		fmt.Println("No se encontraron canciones en", SONGSPATH)
	} else {
		fmt.Println("Se encontraron", len(SUPERPLAYLIST.songs), "canciones")
	}
	fmt.Println("Esperando conexion...")

	// Crear e iniciar el nuevo servidor
	server := NewServer(LISTENER)
	if err := server.Start(); err != nil {
		log.Fatal("Error starting server:", err)
	}

	// El servidor manejará las conexiones entrantes y readLoop en segundo plano.
	fmt.Printf("Server listening on %s\n", LISTENER)

	// Mantener la función main en ejecución para permitir que el servidor maneje las conexiones entrantes.
	select {}
}
