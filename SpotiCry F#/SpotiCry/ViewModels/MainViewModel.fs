namespace SpotiCry.ViewModels

type MainViewModel() =
    inherit ViewModelBase()

    let selectedSongChanged = new Event<_>()
    let mutable selectedSong: string option = None
    let mutable selectedArtist: string option = None
    let mutable selectedLength: string option = None

    member this.SelectedSong 
        with get() = selectedSong
        and set(value) =
            selectedSong <- value
            selectedSongChanged.Trigger((selectedSong, selectedArtist, selectedLength))

    member this.SelectedArtist
        with get() = selectedArtist
        and set(value) =
            selectedArtist <- value
            selectedSongChanged.Trigger((selectedSong, selectedArtist, selectedLength))

    member this.SelectedLength
        with get() = selectedLength
        and set(value) =
            selectedLength <- value
            selectedSongChanged.Trigger((selectedSong, selectedArtist, selectedLength))

    member this.SelectedSongChanged = selectedSongChanged.Publish