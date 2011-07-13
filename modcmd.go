package irc

import (
	"fmt"
	"strings"
)
/*
Defines a module for dynamic loading/unloading of modules from irc commands

To load, use

conn := irc.NewConnection()
conn.SetupModcmd("username1", "username2", "username3")
*/

func (c *IRCConnection) SetupModcmd(admins ...string) {
	handler := func(conn *IRCConnection, chunks []string) {
		if chunks[2] != conn.Info.Channel {
			return
		}

		username := GetUsername(chunks[0])
		var found bool = false

		for _, user := range admins {
			if user == username {
				found = true
				break
			}
		}

		if !found {
			return
		}

		switch strings.Trim(chunks[3], "\r\n") {
		case ":.load":
			for _, word := range chunks[4:] {
				word = strings.Trim(word, "\r\n")
				err := conn.Load(word)
				if err != nil {
					fmt.Fprintln(conn, err)
				} else {
					fmt.Fprintln(conn, "Loaded module:", word)
				}
			}
		case ":.unload":
			for _, word := range chunks[4:] {
				word = strings.Trim(word, "\r\n")
				err := conn.Unload(word)
				if err != nil {
					fmt.Fprintln(conn, err)
				} else {
					fmt.Fprintln(conn, "Unloaded module:", word)
				}
			}
		}
	}

	c.AddCallback("privmsg", handler)
}
