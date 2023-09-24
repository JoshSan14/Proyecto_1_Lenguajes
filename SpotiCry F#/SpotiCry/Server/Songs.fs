module SpotiCry.Server.Songs
open SpotiCry.Server.Connection
open System.Net.Sockets

type Song = {
    Title: string
    Artist: string
    Album: string
    Year: int
    Length: string
}

let getSongsFromServer (client: TcpClient) = // Obtener todas las canciones del servidor
    // Enviar el mensaje para obtener la lista completa de canciones
    sendMessage client "GET_SONG_LIST"
    
    // Recibir la respuesta del servidor
    let response = receiveMessage client

    // Dividir la respuesta en líneas, cada una representando una canción
    let songs = 
        response.Split('\n')
        |> Array.toList
        |> List.choose (fun line ->
            let parts = line.Split("::")
            if parts.Length = 4 then
                let title = parts.[0]
                let artist = parts.[1]
                let year = int parts.[2]
                let length = parts.[3]
                Some { Title = title; Artist = artist; Album = ""; Year = year; Length = length }
            else 
                None)
    // Devolver la lista de canciones del songspath, Ej de 1 dato: Fly Me To The Moon::Frank Sinatra::1964::2:28
    songs