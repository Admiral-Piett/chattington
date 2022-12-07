# chat-telnet
`chat-telnet` is essentially a lightweight chat room server.  Users can connect via telnet protocols, attaching to 
the IP hosting the server and on the port specified in the `app.env` file (which contains all the config).

My approach became a sort of Multiton pattern, where a single server is responsible for fielding connections and 
generating Client objects for each connection.  Clients are held in a memory cache (using the `go-cache` package), 
and are then responsible for managing themselves.  They can, move themselves in and out of different chat rooms, 
change their names, create rooms, even terminate their connection to the server (see full functionality below).

This was fun!  I had a great time dealing with this and had a few more things in mind, but sadly I ran out of time.  
Can't wait to hear what you think!

#### Limitations/Nuances
- You must be in a room for anyone to see what you are writing.
    - When you join you are not in any room at all, which is a valid state but there is no default room.  So 
  messages sent while in there go to no one but you.
- Anything written that is not preceeded by a `\` will be considered a "message" and be sent to anyone in your 
- current room (again, if you are in one).

#### Future Opportunities
- Direct Messaging/Private Rooms - I have some experiments around this on an alternate branch, but I ran out of time.
- API Endpoints or supporting other connection protocols. 

## Installation/Quick Start
Pre-Requisite: Docker (the recommended way to run this is through docker, though you could build and run it as 
a binary too if you like.)

The steps below should use docker to build an image and run the resulting container exposing a port on the 
machine matching the one provided by the PORT environment variable in `app.env`, and automatically tail the logs 
from the server output (example logs below).  When you're finished, a simple `cntrl + c`/`SIGINT` command 
should shutdown the container and then quit the script

- Alter either of the environment variables you wish in `app.env`.
- Run the runner `runner.sh` script.  ex. `sh ./runner.sh`
- Execute `cntrl + c` when finished, to shut down.

Connect to the server with:
- `telnet localhost <PORT>`

## Comands
Upon connecting to the server it should inform you of the available commands (see messaging below).  All commands 
start with the `\` character.  Some have accompanying values, some do not. Commands are not valid unless values 
are allocated correctly.  For most commands (other than those that affect others ie. changing your name, joining 
a room, leaving a room, etc.), other users in your room will not be able to see the output, only you can.
- `\name`: *Accompanying Value Required* - Change your username.  (Upon connection you are given a pseudo-random 
one of the current timestamp)
- `\create`: *Accompanying Value Required* - Create and join a chat room.  Only users in the same room as you are (if any) will ever see any 
messages you send.
- `\join`: *Accompanying Value Required* - Join an existing chat room.
- `\list`: *Accompanying Value Optional* - List members of the specified chat room (if value provided), or list members of the room you're currently in.
- `\leave`: Leave current chat room.
- `\list-rooms`: List all chat rooms and their users.
- `\whoami`: List user information, such as the user's name and current chat room.
- `\exit`: Terminate connection to the chat server.

#### Intro:
```shell
Welcome to Chattington!

Feel free to join any chat rooms you see, or create a room instead, using the available commands below.

Available Commands:
=====
\name 	<user name>		: Change your user name to the <user name> supplied
\create <room name>		: Create and join a new chat room with the <room name> supplied
\join 	<room name>		: Join an existing chat room with the <room name> supplied
\list 	<room name>		: List members in the chat room named after the <room name> supplied
\leave					: Leave the room you are currently in
\list 					: List members in the room you're currently in
\list-rooms				: List all the available rooms and their members
\whoami					: List your name and what room you're currently in
\exit					: Exit server and terminate connection


NOTE: Your user name has been automatically set to `1650575576`
If you'd like to reset it, please use the '\name' command.
```

## Logs
These logs are structured so you can see the time and the relevant user information name of each entry.  Aside 
from the server specific information, each log should indicate the time, the user, and perspective (`>` indicates 
the message sent to the active user, `:` indicates messages sent to the receiving users) of each action.

So each log should be in the format: `<timestamp> <user_name> (> or :) <message>`

#### Examples:
```shell
2022/04/21 21:12:54 Starting chat-telnet server on port: 9000                         // Server start up
2022/04/21 21:12:56 Accepting new connection from address 172.17.0.1:63706            // New telnet connection
2022/04/21 21:13:01 Accepting new connection from address 172.17.0.1:63710
2022/04/21 21:13:10 1650575590: Admiral> User: 1650575576 has become -> Admiral       // User changing names
2022/04/21 21:13:24 1650575604: Captain> User: 1650575581 has become -> Captain
2022/04/21 21:13:37 Creating Room: boat-room                                          // User creating room
2022/04/21 21:13:37 1650575617: Captain> New room created: boat-room
2022/04/21 21:13:44 1650575624: Admiral: Admiral has entered: boat-room               // User joining room
2022/04/21 21:13:44 1650575624: Admiral> Admiral has entered: boat-room
2022/04/21 21:13:48 1650575628: Admiral: Ahoy!                                        // Chat messages
2022/04/21 21:13:48 1650575628: Admiral> Ahoy!
2022/04/21 21:14:00 1650575640: Captain> Ah, it's you Matey!
2022/04/21 21:14:00 1650575640: Captain: Ah, it's you Matey!
2022/04/21 21:14:04 Removed connection 172.17.0.1:63706 from pool                     // Breaking their connection
2022/04/21 21:14:04 1650575644: Admiral: Admiral has gone offline
2022/04/21 21:14:04 1650575644: Admiral> Admiral has gone offline
2022/04/21 21:14:08 Removed connection 172.17.0.1:63710 from pool
2022/04/21 21:14:08 1650575648: Captain> Captain has gone offline
2022/04/21 21:14:08 1650575648: Captain: Captain has gone offline
^Cstopping chat                                                                         // Shutting down the server
```

## Tests
Feel free to run any and all unit tests with `go test ./...`.
