package irc

/*
Handles callback adding/deleting/checking for a connection
*/

//Adds a callback to the handler
func (conn *Connection) AddCallback(cmd string, call Callback) {
	conn.callbacks[cmd] = append(conn.callbacks[cmd], call)
}
