
// TcpClientObject.cpp
#include "TcpClientObject.h"
#include "TcpSocket.h"

bool bIsPlayerConnected = false;

UTcpClientObject::UTcpClientObject()
    : TcpSocket(nullptr)
    , bIsPlayerConnected(false)
{
}


bool UTcpClientObject::ConnectToServer(const FString& ServerAddress, int32 ServerPort)
{
    SavedServerAddress = ServerAddress;
    SavedServerPort = ServerPort;
    TcpSocket = new FTcpSocket();
    bool bIsConnected = TcpSocket->Connect(ServerAddress, ServerPort);
    if (bIsConnected)
    {
        this->bIsPlayerConnected = true;
        OnConnectedToServer.Broadcast();
        StartHeartbeatTimer();
    }
    else
    {
        UE_LOG(LogTemp, Error, TEXT("Failed to connect to server: %s:%d"), *ServerAddress, ServerPort);
        // Try again in 5 seconds
        GetWorld()->GetTimerManager().SetTimer(ReconnectTimerHandle, this, &UTcpClientObject::RetryConnectToServer, 5.0f, false);
    }
    return bIsConnected;
}

void UTcpClientObject::RetryConnectToServer()
{
    if (!bIsPlayerConnected)
    {
        UE_LOG(LogTemp, Warning, TEXT("Retrying connection to server..."));
        ConnectToServer(SavedServerAddress, SavedServerPort);
    }
}



void UTcpClientObject::DisconnectFromServer()
{
    StopHeartbeatTimer();

    if (TcpSocket != nullptr)
    {
        TcpSocket->Disconnect();
        delete TcpSocket;
        TcpSocket = nullptr;
    }
}


void UTcpClientObject::DisconnectAndResetConnection()
{
    StopHeartbeatTimer();

    if (TcpSocket != nullptr)
    {
        TcpSocket->Disconnect();
        delete TcpSocket;
        TcpSocket = nullptr;
    }
    bIsPlayerConnected = false;
}

bool UTcpClientObject::ReconnectToServer(const FString& ServerAddress, int32 ServerPort)
{
    if (bIsPlayerConnected) // Check if the player is already connected
    {
        UE_LOG(LogTemp, Warning, TEXT("Player is already connected"));
        return true;
    }
    DisconnectAndResetConnection(); // Disconnect the client and reset the flag
    return ConnectToServer(ServerAddress, ServerPort); // Attempt to connect to the server again
}

bool UTcpClientObject::IsPlayerConnected() const
{
    return bIsPlayerConnected;
}

void UTcpClientObject::StopHeartbeatTimer()
{
    GetWorld()->GetTimerManager().ClearTimer(HeartbeatTimerHandle);
}

void UTcpClientObject::StartHeartbeatTimer()
{
    GetWorld()->GetTimerManager().SetTimer(HeartbeatTimerHandle, this, &UTcpClientObject::SendPing, 5.0f, true);
}

void UTcpClientObject::SendPing()
{
    SendEvent(TEXT("PING"), TEXT(""));
}


bool UTcpClientObject::SendEvent(const FString& EventName, const FString& EventBody)
{
    if (!bIsPlayerConnected)
    {
        UE_LOG(LogTemp, Warning, TEXT("Player is not connected"));
        return false;
    }
    if (TcpSocket == nullptr)
    {
        UE_LOG(LogTemp, Error, TEXT("TcpSocket is null"));
        return false;
    }
    bool bResult = TcpSocket->SendEvent(EventName, EventBody);
    if (!bResult)
    {
        UE_LOG(LogTemp, Error, TEXT("Failed to send event: %s"), *EventName);
        // Try again in 1 second
        GetWorld()->GetTimerManager().SetTimer(RetrySendTimerHandle, this, &UTcpClientObject::RetrySendEvent, 1.0f, false, 1.0f);
    }
    return bResult;
}

void UTcpClientObject::RetrySendEvent()
{
    UE_LOG(LogTemp, Warning, TEXT("Retrying to send event..."));
    FString EventName, EventBody;
    TcpSocket->GetLastEvent(EventName, EventBody);
    SendEvent(EventName, EventBody);
}

void UTcpClientObject::SendMoveCommand(int32 PlayerControllerId, float X, float Y)
{
    // Send the MOVE command to the server
    FString EventBody = FString::Printf(TEXT("%d,%f,%f"), PlayerControllerId, X, Y);
    SendEvent(TEXT("MOVE"), EventBody);
}

void UTcpClientObject::SpawnObject(int32 ObjectID, float X, float Y)
{
    // Search for the object data in the DataTable
    UDataTable* ObjectDataTable = LoadObject<UDataTable>(nullptr, TEXT("DataTable'/Game/Data/ObjectDataTable.ObjectDataTable'"));
    if (ObjectDataTable == nullptr)
    {
        UE_LOG(LogTemp, Error, TEXT("ObjectDataTable not found"));
        return;
    }

    const FString ContextString = FString(TEXT("ObjectData"));
    const FObjectDataTableRow* ObjectData = ObjectDataTable->FindRow<FObjectDataTableRow>(FName(*FString::FromInt(ObjectID)), ContextString);
    if (ObjectData == nullptr)
    {
        UE_LOG(LogTemp, Error, TEXT("ObjectData not found for ID %d"), ObjectID);
        return;
    }

    // Spawn the object at the specified location
    AActor* Object = GetWorld()->SpawnActor(ObjectData->ObjectClass, FVector(X, Y, 0), FRotator::ZeroRotator);
    if (Object == nullptr)
    {
        UE_LOG(LogTemp, Error, TEXT("Failed to spawn object"));
        return;
    }
}

void UTcpClientObject::OnPacketReceived(const FString& EventName, const FString& EventBody)
{
    if (EventName == TEXT("MOVE"))
    {
        // Handle MOVE event
        OnMovePlayer(EventBody);
    }
    else if (EventName == TEXT("SPAWN_PLAYER"))
    {
        // Handle SPAWN_PLAYER event
        OnSpawnPlayer(EventBody);
    }
    else if (EventName == TEXT("CHAT"))
    {
        OnChatMessage(EventBody);
    }
    else if (EventName == TEXT("PRIVATE_CHAT"))
    {
        OnPrivateChatMessage(EventBody);
    }
    else if (EventName == TEXT("SPAWN_OBJECT"))
    {
        // Handle SPAWN_OBJECT event
        TArray<FString> EventData;
        EventBody.ParseIntoArray(EventData, TEXT(","));

        if (EventData.Num() == 3)
        {
            int32 ObjectID = FCString::Atoi(*EventData[0]);
            float X = FCString::Atof(*EventData[1]);
            float Y = FCString::Atof(*EventData[2]);

            SpawnObject(ObjectID, X, Y);
        }
    }
}


void UTcpClientObject::OnSpawnPlayer(const FString& EventBody)
{
    // Parse the event body
    TArray<FString> EventData;
    EventBody.ParseIntoArray(EventData, TEXT(","));

    if (EventData.Num() == 3)
    {
        // Get the player controller id, x and y coordinates from the event data
        int32 PlayerControllerId = FCString::Atoi(*EventData[0]);
        float X = FCString::Atof(*EventData[1]);
        float Y = FCString::Atof(*EventData[2]);

        // Spawn the player controller and pawn
        APlayerController* PlayerController = UGameplayStatics::GetPlayerController(GetWorld(), PlayerControllerId);
        if (PlayerController == nullptr)
        {
            // Spawn a new player controller
            PlayerController = UGameplayStatics::SpawnPlayerController(GetWorld(), PlayerControllerId);
        }

        // Spawn the pawn for the player controller
        APawn* Pawn = GetWorld()->SpawnActor<APawn>(APawn::StaticClass(), FVector(X, Y, 0), FRotator::ZeroRotator);
        if (Pawn != nullptr)
        {
            PlayerController->Possess(Pawn);
        }
    }
}

void UTcpClientObject::OnMovePlayer(const FString& EventBody)
{
    TArray<FString> EventData;
        EventBody.ParseIntoArray(EventData, TEXT(","));

        if (EventData.Num() == 3)
        {
            // Get the player controller id, x and y coordinates from the event data
            int32 PlayerControllerId = FCString::Atoi(*EventData[0]);
            float X = FCString::Atof(*EventData[1]);
            float Y = FCString::Atof(*EventData[2]);

            // Move the player controller's pawn to the specified location
            APlayerController* PlayerController = UGameplayStatics::GetPlayerController(GetWorld(), PlayerControllerId);
            if (PlayerController != nullptr)
            {
                APawn* Pawn = PlayerController->GetPawn();
                if (Pawn != nullptr)
                {
                    Pawn->SetActorLocation(FVector(X, Y, Pawn->GetActorLocation().Z));
                }
            }
        }
}

bool UTcpClientObject::SendChatMessage(const FString& SenderName, const FString& Message)
{
    if (!bIsPlayerConnected) // Check if the player is still connected
    {
        UE_LOG(LogTemp, Warning, TEXT("Player is not connected"));
        return false;
    }
    if (TcpSocket == nullptr)
    {
        return false;
    }

    FString EventBody = FString::Printf(TEXT("%s,%s"), *SenderName, *Message);
    return TcpSocket->SendEvent(TEXT("CHAT"), EventBody);
}

void UTcpClientObject::OnChatMessage(const FString& EventBody)
{
    TArray<FString> EventData;
    EventBody.ParseIntoArray(EventData, TEXT(","));

    if (EventData.Num() == 2)
    {
        FString SenderName = EventData[0];
        FString Message = EventData[1];

        OnChatMessageReceived.Broadcast(SenderName, Message);
    }
}

void UTcpClientObject::Tick(float DeltaTime)
{
    Super::Tick(DeltaTime);

    ReceivePackets();
}

void UTcpClientObject::ReceivePackets()
{
    if (!bIsPlayerConnected) // Check if the player is still connected
    {
        return;
    }
    if (TcpSocket == nullptr)
    {
        return;
    }

    FString PacketData;
    while (TcpSocket->ReceivePacket(PacketData))
    {
        PacketBuffer += PacketData;

        FString EventName, EventBody;
        if (TcpSocket->ParsePacket(PacketBuffer, EventName, EventBody))
        {
            OnPacketReceived(EventName, EventBody);
        }
    }
}

bool UTcpClientObject::SendPrivateChatMessage(const FString& SenderName, const FString& RecipientName, const FString& Message)
{
    if (!bIsPlayerConnected) // Check if the player is still connected
    {
        UE_LOG(LogTemp, Warning, TEXT("Player is not connected"));
        return false;
    }
    if (TcpSocket == nullptr)
    {
        return false;
    }

    FString EventBody = FString::Printf(TEXT("%s,%s"), *RecipientName, *Message);
    return TcpSocket->SendEvent(TEXT("PRIVATE_CHAT"), EventBody);
}

void UTcpClientObject::OnPrivateChatMessage(const FString& EventBody)
{
    TArray<FString> EventData;
    EventBody.ParseIntoArray(EventData, TEXT(","));

    if (EventData.Num() == 2)
    {
        FString SenderName = EventData[0];
        FString Message = EventData[1];

        OnPrivateChatMessageReceived.Broadcast(SenderName, Message);
    }
}