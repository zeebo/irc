package irc

//Adds a callback to the handler
func (conn *IRCConnection) AddCallback(cmd string, call Callback) {
	conn.callbacks[cmd] = append(conn.callbacks[cmd], call)
}

//Removes a callback from the handler
func (conn *IRCConnection) DelCallback(cmd string, call Callback) {
	calls, exists := conn.callbacks[cmd]
	if !exists {
		return
	}

	for idx, other := range calls {
		if other == call {
			conn.callbacks[cmd] = append(calls[:idx], calls[idx+1:]...)
			break
		}
	}
}

//Checks if a specified callback exists
func (conn *IRCConnection) HasCallback(cmd string, call Callback) (ret bool) {
	calls, exists := conn.callbacks[cmd]
	if !exists {
		return false
	}

	for _, other := range calls {
		if other == call {
			return true
		}
	}

	return false
}
