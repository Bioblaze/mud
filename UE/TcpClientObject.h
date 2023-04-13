#pragma once

#include "CoreMinimal.h"
#include "UObject/NoExportTypes.h"
#include "TcpSocket.h"
#include "TcpClientObject.generated.h"

DECLARE_DYNAMIC_MULTICAST_DELEGATE(FOnConnectedToServer);
DECLARE_DYNAMIC_MULTICAST_DELEGATE_TwoParams(FOnChatMessageReceived, const FString&, SenderName, const FString&, Message);
DECLARE_DYNAMIC_MULTICAST_DELEGATE_TwoParams(FOnPrivateChatMessageReceived, const FString&, SenderName, const FString&, Message);

UCLASS()
class MYPROJECT_API UTcpClientObject : public UObject
{
    GENERATED_BODY()

public:
    UTcpClientObject();

    UFUNCTION(BlueprintCallable, Category = "TcpClient")
    bool ConnectToServer(const FString& ServerAddress, int32 ServerPort);
    UFUNCTION(BlueprintCallable, Category = "TcpClient")
    void RetryConnectToServer();
    UFUNCTION(BlueprintCallable, Category = "TcpClient")
    void DisconnectFromServer();
    UFUNCTION(BlueprintCallable, Category = "TcpClient")
    void DisconnectAndResetConnection();
    UFUNCTION(BlueprintCallable, Category = "TcpClient")
    bool ReconnectToServer(const FString& ServerAddress, int32 ServerPort);
    UFUNCTION(BlueprintCallable, Category = "TcpClient")
    bool IsPlayerConnected() const;
    UFUNCTION(BlueprintCallable, Category = "TcpClient")
    bool SendEvent(const FString& EventName, const FString& EventBody);
    UFUNCTION(BlueprintCallable, Category = "TcpClient")
    void RetrySendEvent();
    UFUNCTION(BlueprintCallable, Category = "TcpClient")
    void SendMoveCommand(int32 PlayerControllerId, float X, float Y);
    UFUNCTION(BlueprintCallable, Category = "TcpClient")
    void SpawnObject(int32 ObjectID, float X, float Y);
    UFUNCTION(BlueprintCallable, Category = "TcpClient")
    void OnPacketReceived(const FString& EventName, const FString& EventBody);
    UFUNCTION(BlueprintCallable, Category = "TcpClient")
    void OnSpawnPlayer(const FString& EventBody);
    UFUNCTION(BlueprintCallable, Category = "TcpClient")
    void OnMovePlayer(const FString& EventBody);
    UFUNCTION(BlueprintCallable, Category = "TcpClient")
    bool SendChatMessage(const FString& SenderName, const FString& Message);
    UFUNCTION(BlueprintCallable, Category = "TcpClient")
    void OnChatMessage(const FString& EventBody);
    void Tick(float DeltaTime);
    UFUNCTION(BlueprintCallable, Category = "TcpClient")
    void ReceivePackets();
    UFUNCTION(BlueprintCallable, Category = "TcpClient")
    bool SendPrivateChatMessage(const FString& SenderName, const FString& RecipientName, const FString& Message);
    UFUNCTION(BlueprintCallable, Category = "TcpClient")
    void OnPrivateChatMessage(const FString& EventBody);

    UPROPERTY(BlueprintAssignable, Category = "TcpClient")
    FOnConnectedToServer OnConnectedToServer;

    UPROPERTY(BlueprintAssignable, Category = "TcpClient")
    FOnChatMessageReceived OnChatMessageReceived;

    UPROPERTY(BlueprintAssignable, Category = "TcpClient")
    FOnPrivateChatMessageReceived OnPrivateChatMessageReceived;

private:
    FTcpSocket* TcpSocket;
    bool bIsPlayerConnected;
    FString SavedServerAddress;
    int32 SavedServerPort;
    FTimerHandle ReconnectTimerHandle;
    FTimerHandle HeartbeatTimerHandle;
    FTimerHandle RetrySendTimerHandle;

    void StopHeartbeatTimer();
    void StartHeartbeatTimer();
    void SendPing();
};
