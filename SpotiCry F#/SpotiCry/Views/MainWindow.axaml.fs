namespace SpotiCry.Views

open Avalonia
open Avalonia.Controls
open SpotiCry.ViewModels
open Avalonia.Controls.Shapes
open Avalonia.Media
open Avalonia.Markup.Xaml
open NAudio.Wave
open System
open System.IO
open SpotiCry.Server.Connection
open System.Net.Sockets
open Avalonia.Threading

type MainWindow() as this = 
    inherit Window ()
    let mutable clientConnection: TcpClient option = None
    let viewModel = MainViewModel()  // Crear una instancia del ViewModel

    //audio
    let mutable currentlyPlayingSong: string option = None // Para comparar con el selectedSong cada vez que se presiona Play, para evitar reproducir la misma canción y manejar la lógica de reproducir otra seleccionada
    let mutable isPlaying = false
    let mutable isPaused = false
    let mutable waveOut = new WaveOutEvent()  // Una única instancia de WaveOutEvent
    let mutable timer = new DispatcherTimer()

    do
       this.DataContext <- viewModel
       this.InitializeComponent()
       this.refreshConnection()
       this.addEventHandlers()
       this.setupViewModelSubscriptions()

       let songsView = this.FindControl<SongsView>("SongsView")
       songsView.DataContext <- viewModel

    member private this.InitializeComponent() =
#if DEBUG
        this.AttachDevTools()
#endif
        AvaloniaXamlLoader.Load(this)

    // Funcion para agregar manejadores de eventos
    member private this.addEventHandlers() =
        let refreshButton = this.FindControl<Button>("RefreshButton")
        refreshButton.Click.Add(this.handleRefresh)

        // manejador de eventos para el boton "Play"
        let playButton = this.FindControl<Button>("PlayButton")
        playButton.Click.Add(this.handlePlay)

        //volume slider
        let volumeSlider: Slider = this.FindControl<Slider>("VolumeSlider")
        volumeSlider.Value <- 50.0
        volumeSlider.ValueChanged.Add(this.handleVolumeChange)

        //Filter buttons
        let yearButton = this.FindControl<Button>("yearFLT")
        yearButton.Click.Add(this.handleOrderByYear)

        (*let lengthButton = this.FindControl<Button>("lengthFLT")
        lengthButton.Click.Add(this.handleOrderByLength) *)

        let nameButton = this.FindControl<Button>("nameFLT")
        nameButton.Click.Add(this.handleOrderByName)

    // Funcion que maneja el evento de clic en el boton "Refresh"
    member private this.handleRefresh _ =  
        this.refreshConnection()

    // Actualizar el slider de progreso
    member private this.updateSlider mp3Reader =
        timer.Stop()  // Detiene el temporizador anterior

        let slider: Slider = this.FindControl<Slider>("ProgressBar")
        let totalLength = (mp3Reader : Mp3FileReader).TotalTime.TotalSeconds
        slider.Maximum <- totalLength

        timer <- new DispatcherTimer()
        timer.Interval <- TimeSpan.FromMilliseconds(1000.0) // Actualiza cada 1s

        timer.Tick.Add (fun _ ->
            if not isPaused then
                let currentPosition = (mp3Reader : Mp3FileReader).CurrentTime.TotalSeconds
                slider.Value <- currentPosition
        )
        timer.Start()

    // Funcion para el slider de volumen
    member private this.handleVolumeChange (args: Primitives.RangeBaseValueChangedEventArgs) =
        let newVolume = args.NewValue
        waveOut.Volume <- (float32 newVolume) / 100.0f

    member private this.handleOrderByYear _ =
        // Obtener referencia a SongsView y reordenar el panel de canciones por año
        let songsView = this.FindControl<SongsView>("SongsView")
        songsView.OrderByYear()

    (*member private this.handleOrderByLength _ =
        // Obtener referencia a SongsView y reordenar el panel de canciones por duración
        let songsView = this.FindControl<SongsView>("SongsView")
        songsView.OrderByLength() *)

    member private this.handleOrderByName _ =
        // Obtener referencia a SongsView y reordenar el panel de canciones por nombre
        let songsView = this.FindControl<SongsView>("SongsView")
        songsView.OrderByName()


    // Funcion que refresca la conexi�n y actualiza el indicador
    member private this.refreshConnection() =
        let statusIndicator = this.FindControl<Ellipse>("StatusIndicator")

        // Funcion para actualizar el indicador en la UI
        let updateUI () =
            match clientConnection with
            | Some client when isConnected client -> 
                statusIndicator.Fill <- SolidColorBrush(Colors.Green)
            
                // Obtener referencia a SongsView y llenar el panel de canciones
                let songsView = this.FindControl<SongsView>("SongsView")
                songsView.FillSongsPanel client
            
            | _ -> 
                statusIndicator.Fill <- SolidColorBrush(Colors.Red)

        // Intentar establecer la conexi�n de forma as�ncrona
        async {
            match clientConnection with
            | Some client when not (isConnected client) -> 
                do! Async.SwitchToThreadPool()
                this.initializeConnection()
                Dispatcher.UIThread.Post(updateUI)
            | None -> 
                do! Async.SwitchToThreadPool()
                this.initializeConnection()
                Dispatcher.UIThread.Post(updateUI)
            | _ -> 
                Dispatcher.UIThread.Post(updateUI)
        } |> Async.StartImmediate


    // Funci�n para intentar conectarse al servidor y actualizar el indicador
    member private this.initializeConnection() =
        try
            match tryConnectToServer () with
            | Some client ->
                clientConnection <- Some client
            | None -> ()
        with
        | e ->
            printfn "Error: %s" e.Message
            
    // Configurar suscripciones al ViewModel
    member private this.setupViewModelSubscriptions() =
        //printfn "El DataContext en MainWindow es: %A" viewModel
        // Observar cambios en SelectedSong
        viewModel.SelectedSongChanged.Add(fun (newSong, newArtist, newLength) -> 
            let songInfoText = this.FindControl<TextBlock>("SongInfoText")
            let songLengthText = this.FindControl<TextBlock>("SongLengthText")
    
            let title = match newSong with
                        | Some t -> t
                        | None -> "No song"
            let artist = match newArtist with
                         | Some a -> a
                         | None -> "Unknown artist"
            let length = match newLength with
                         | Some l -> l
                         | None -> "0:00"
    
            songInfoText.Text <- sprintf "%s - %s" title artist
            songLengthText.Text <- length
        )

    // Manejador de eventos para el botón "Play"
    member private this.handlePlay _ =
        async {
            try
                match viewModel.SelectedSong with
                | Some songTitle ->
                    if currentlyPlayingSong <> Some songTitle then  // Verifica si se ha cambiado la canción
                        waveOut.Stop()  // Detiene la reproducción actual
                        isPlaying <- false  // Actualiza el estado de reproducción
                        currentlyPlayingSong <- Some songTitle  // Actualiza la canción actual

                        let progressBar: Slider = this.FindControl<Slider>("ProgressBar")
                        progressBar.Value <- 0.0  // Restablece el valor del deslizador
                        timer.Stop()  // Detiene el temporizador

                    if not isPlaying then  // Verificar si ya está sonando
                        let messageToSend = sprintf "PLAY_S::SUPERPLAYLIST::%s::NONE" songTitle
                        match clientConnection with
                        | Some client ->
                            sendMessage client messageToSend
                            printfn "Mensaje enviado al servidor: %s" messageToSend

                            let receivedBytes = receiveBytes client
                            let memoryStream = new MemoryStream(receivedBytes)

                            waveOut.Dispose()
                            waveOut <- new WaveOutEvent()

                            let mp3Reader = new Mp3FileReader(memoryStream)
                            this.updateSlider mp3Reader
                            waveOut.Init(mp3Reader)
                            waveOut.Play()
                            isPlaying <- true
                            isPaused <- false

                        | None -> 
                            printfn "No se pudo enviar el mensaje, no hay una conexión activa."
                    else
                        if not isPaused then
                            waveOut.Pause()
                            isPaused <- true
                        else
                            waveOut.Play()
                            isPaused <- false

                | None -> ()
            with
            | e ->
                printfn "Se produjo un error durante la reproducción: %s" e.Message
        }
        |> Async.StartImmediate