# login server

This program is a client server utilizing unix domain sockets (uds) for communication.
The server manages a kv database (https://github.com/akrylysov/pogreb).

Features:
 - graceful shutdown due to interruptions
 - go handlers for each connection

## loginDbServerV3.go

The server program operates until shutdown.

## dbadmin

The client program runs in a loop to interact with the server and the database.
