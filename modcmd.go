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

func (c *Connection) SetupModcmd(admins ...string) {
	handler := func(conn *Connection, chunks []string) {
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
		case ":.info":
			for _, word := range chunks[4:] {
				word = strings.Trim(word, "\r\n")
				module, exists := conn.modules[word]
				if !exists {
					fmt.Fprintln(conn, "Module does not exist:", word)
				} else {
					fmt.Fprintf(conn, "%s: %s\n", word, module.Info)
				}
			}
		case ":.loaded":
			fmt.Fprint(conn, "Loaded modules: ")
			for _, module := range conn.modules {
				if module.loaded {
					fmt.Fprint(conn, module.Name, " ")
				}
			}
			fmt.Fprint(conn, "\n")
		case ":.list":
			fmt.Fprint(conn, "Modules: ")
			for _, module := range conn.modules {
				fmt.Fprint(conn, module.Name, " ")
			}
			fmt.Fprint(conn, "\n")
		}
	}

	c.AddCallback("privmsg", handler)
}
