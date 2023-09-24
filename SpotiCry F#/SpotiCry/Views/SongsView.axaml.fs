namespace SpotiCry.Views

open Avalonia.Controls
open SpotiCry.ViewModels
open Avalonia.Markup.Xaml
open SpotiCry.Server.Songs
open System.Net.Sockets
open System.Linq

type SongsView () as this = 
    inherit UserControl ()
    
    let buttonsList = new System.Collections.Generic.List<Button>() // Lista para almacenar los botones
    let mutable songsStackPanel: StackPanel = null

    do 
        this.InitializeComponent()
        songsStackPanel <- this.FindControl<StackPanel>("SongsStackPanel")
    
    // Funci�n para llenar el StackPanel con botones
    member internal this.FillSongsPanel (client: TcpClient) =
        // Obtener la lista de canciones del servidor
        let songs = getSongsFromServer client

        // Limpiar cualquier boton existente
        songsStackPanel.Children.Clear()
        buttonsList.Clear()

        // Crear un boton para cada cancion y añadirlo al StackPanel
        for song in songs do
            let btn = new Button()
            btn.Content <- sprintf "%s - %s" song.Title song.Artist
            // Almacenar year y length para usarlos en los filters
            btn.Tag <- (song.Year, song.Length)

            // Agregar un manejador de eventos para actualizar selectedSong
            btn.Click.Add(fun _ ->
                let viewModel = this.DataContext :?> MainViewModel  // Obtener la instancia compartida
                viewModel.SelectedSong <- Some song.Title
                viewModel.SelectedArtist <- Some song.Artist

                // Convertir la duración al formato deseado antes de asignarla a SelectedLength
                let originalFormat = song.Length
                let mins = originalFormat.Split('m').[0]
                let secs = originalFormat.Split('m').[1].TrimEnd('s')
                let paddedSecs = if secs.Length = 1 then "0" + secs else secs  // Asegurarse de que los segundos tengan dos dígitos
                let formattedDuration = sprintf "%s:%s" mins paddedSecs

                viewModel.SelectedLength <- Some formattedDuration
            )

            // Establecer las propiedades visuales
            btn.Background <- new Avalonia.Media.SolidColorBrush(Avalonia.Media.Colors.Transparent)
            btn.HorizontalAlignment <- Avalonia.Layout.HorizontalAlignment.Stretch
            btn.BorderBrush <- new Avalonia.Media.SolidColorBrush(Avalonia.Media.Colors.DarkGray)
            btn.BorderThickness <- new Avalonia.Thickness(1.0)
            btn.FontSize <- 15.0
    
            // Añadir el botón al StackPanel
            buttonsList.Add(btn)

        // Añadir los botones al StackPanel desde la lista
        for btn in buttonsList do
            songsStackPanel.Children.Add(btn)


    // Métodos para ordenar los botones
    member internal this.OrderByYear() =
        do
            let orderedList = buttonsList.OrderBy(fun btn -> btn.Tag.ToString().Split('|').[0]).ToList()
            this.UpdateSongsPanel(orderedList)

   (* member internal this.OrderByLength() =
        let orderedList = 
            buttonsList
            |> List.sortBy (fun btn ->
                let originalLength = btn.Tag.ToString().Split('|').[1]
                let mins = int originalLength.Split(':').[0] 
                let secs = int originalLength.Split(':').[1]
                mins * 60 + secs) 
        this.UpdateSongsPanel(orderedList) *)

    member internal this.OrderByName() =
        do
            let orderedList = buttonsList.OrderBy(fun btn -> btn.Content.ToString()).ToList()
            this.UpdateSongsPanel(orderedList)

    member internal this.UpdateSongsPanel(orderedList) =
        do
            songsStackPanel.Children.Clear()
            for btn in orderedList do
                songsStackPanel.Children.Add btn

    member private this.InitializeComponent() =
        AvaloniaXamlLoader.Load(this)