namespace SpotiCry

open Avalonia
open Avalonia.Controls.ApplicationLifetimes
open Avalonia.Markup.Xaml
open SpotiCry.ViewModels
open SpotiCry.Views

type App() =
    inherit Application()

    override this.Initialize() =
            AvaloniaXamlLoader.Load(this)

    override this.OnFrameworkInitializationCompleted() =
        match this.ApplicationLifetime with
        | :? IClassicDesktopStyleApplicationLifetime as desktopLifetime ->
            desktopLifetime.MainWindow <- MainWindow(DataContext = MainViewModel())
        | :? ISingleViewApplicationLifetime as singleViewLifetime ->
            singleViewLifetime.MainView <- SongsView(DataContext = MainViewModel())
        | _ -> ()

        base.OnFrameworkInitializationCompleted()
