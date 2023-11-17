module turboGuac/server

go 1.21.3

require (
	nhooyr.io/websocket v1.8.10
	turboGuac/message v0.0.0-00010101000000-000000000000
)

replace turboGuac/message => ../message
