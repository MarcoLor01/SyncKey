@echo off
setlocal

REM Imposta la variabile di ambiente CONFIG

REM Percorsi dei file Go
set "server_files=servers"
set "client_files=clients"

REM Set CONFIG from command line argument
set CONFIG=%1

if %CONFIG% == 1 (
    REM Esecuzione dei server in nuovi terminali cmd.exe
    set ENV_PATH=../.env
    start cmd.exe /K "cd %server_files% && go run -race server1.go -port=1234"
    REM start cmd.exe /K "cd servers && dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient server2.go
    start cmd.exe /K "cd %server_files% && go run -race server2.go -port=2345"
    start cmd.exe /K "cd %server_files% && go run -race server3.go -port=3456"

    REM Esecuzione del client in un nuovo terminale cmd.exe
    start cmd.exe /K "cd %client_files% && go run -race clientTests.go"
) else if %CONFIG% == 2 (
    start cmd.exe /K "docker compose up --build"
) else (
    echo Errore: CONFIG deve essere 1 o 2.
    exit /b 1
)

endlocal