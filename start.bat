@echo off
setlocal

REM Imposta la variabile di ambiente CONFIG

REM Percorsi dei file Go
set "server_files=servers"
set "client_files=clients"
set CONFIG = 0
set "operation_type=seq"
REM Verifica se Go è installato
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo Errore: Go non trovato nel sistema.
    exit /b 1
)

REM Esecuzione dei server in nuovi terminali cmd.exe
  start cmd.exe /K "cd %server_files% && go run server1.go -m %operation_type%"
  start cmd.exe /K "cd %server_files% && go run server2.go -m %operation_type%"
  start cmd.exe /K "cd %server_files% && go run server3.go -m %operation_type%"
  start cmd.exe /K "cd %server_files% && go run server4.go -m %operation_type%"
  start cmd.exe /K "cd %server_files% && go run server5.go -m %operation_type%"

REM Esecuzione del client in un nuovo terminale cmd.exe
start cmd.exe /K "cd %client_files% && go run client.go"


endlocal

  