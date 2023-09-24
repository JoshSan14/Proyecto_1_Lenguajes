module SpotiCry.Server.Connection

open System.Net.Sockets
open System.Text
open System.IO

let serverIP = "127.0.0.1" //localhost
let serverPort = 8081     //puerto del servidor

let tryConnectToServer () =
    try
        let client = new TcpClient(serverIP, serverPort)
        Some client
    with
    | :? SocketException -> None

let isConnected (client: TcpClient) = //Función para verificar si el cliente sigue conectado, manda un ping al server y si no es posible enviar el mensaje se asume que el cliente no está conectado.
    try
        if client.Connected then
            let stream = client.GetStream()
            let buffer = Encoding.ASCII.GetBytes("ping")
            stream.Write(buffer, 0, buffer.Length)
            true
        else
            false
    with
    | _ -> false

let sendMessage (client: TcpClient) message =
    let stream = client.GetStream()
    let messageBytes = Encoding.ASCII.GetBytes(message : string)
    stream.Write(messageBytes, 0, messageBytes.Length)

let receiveMessage (client: TcpClient) =
    let buffer = Array.zeroCreate 5000  // o cualquier otro tamaño que hayas configurado
    let stream = client.GetStream()
    let bytesRead = stream.Read(buffer, 0, buffer.Length)
    
    // Imprimir la longitud del buffer y la cantidad de bytes leídos
    printfn "Longitud del buffer: %d" buffer.Length
    printfn "Bytes leídos: %d" bytesRead
    
    Encoding.ASCII.GetString(buffer, 0, bytesRead)

let receiveBytes (client: TcpClient) =
    let buffer = Array.zeroCreate 10000000  // Buffer para almacenar los bytes recibidos
    let stream = client.GetStream()
    use memoryStream = new MemoryStream()

    let rec loop() =
        let bytesRead = stream.Read(buffer, 0, buffer.Length)
        if bytesRead > 0 then
            memoryStream.Write(buffer, 0, bytesRead)
        let receivedData = Encoding.ASCII.GetString(buffer, 0, bytesRead)
        if receivedData.Contains("EOF") then
            memoryStream.SetLength(memoryStream.Length - 3L) // Eliminar los 3 últimos bytes ("EOF"), se usa solo para confirmación de haber recibido todo el archivo)
            printfn "Transmisión completa. Bytes recibidos: %d" memoryStream.Length
        else
            loop()

    loop()
    memoryStream.ToArray()