#!/bin/bash


server_files="servers"
client_files="clients"

CONFIG=$1

if [ $CONFIG -eq 1 ]; then
    # Esecuzione dei server in nuovi terminali
    ENV_PATH="../.env"
    gnome-terminal -- bash -c "cd $server_files && go run server1.go"
    gnome-terminal -- bash -c "cd $server_files && go run server2.go"
    gnome-terminal -- bash -c "cd $server_files && go run server3.go"

    # Esecuzione del client in un nuovo terminale
    gnome-terminal -- bash -c "cd $client_files && go run client.go"
elif [ $CONFIG -eq 2 ]; then
    # Esecuzione di docker compose up --build
    docker-compose up --build
else
    echo "Errore: CONFIG deve essere 1 o 2."
    exit 1
fi
