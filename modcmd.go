package irc

import (
	"fmt"
	"strings"
)

/*
Sets up a callback for dynamic loading/unloading of modules from irc commands

To load, use

conn := irc.NewConnection()
conn.SetupModcmd("username1", "username2", "username3")

Then in irc, you can do things like

.load <module1> [module2] [module3] ... [moduleN]
.unload <module1> [module2] [module3] ... [moduleN]
.info <module1> [module2] [module3] ... [moduleN]
.list
.loaded

to control and inspect module behavior
*/

//Sets up the callback for loading and unloading. Don't call this twice
//for any admin. You'll probably get some wacky behaviors.
func (c *Connection) SetupModcmd(admins ...string) {
	//Create a function that closes on the value of admins
	handler := func(conn *Connection, chunks []string) {
		if chunks[2] != conn.Info.Channel {
			return
		}

		//Make sure the user is an admin
		username := GetUsername(chunks[0])
		var found bool = false

		for _, user := range admins {
			if user == username {
				found = true
				break
			}
		}

		//bail out if they arent found
		if !found {
			return
		}

		//switch on the command they sent
		switch chunks[3] {
		case ".load":
			for _, word := range chunks[4:] {
				word = strings.Trim(word, "\r\n")
				err := conn.Load(word)
				if err != nil {
					fmt.Fprintln(conn, err)
				} else {
					fmt.Fprintln(conn, "Loaded module:", word)
				}
			}
		case ".unload":
			for _, word := range chunks[4:] {
				word = strings.Trim(word, "\r\n")
				err := conn.Unload(word)
				if err != nil {
					fmt.Fprintln(conn, err)
				} else {
					fmt.Fprintln(conn, "Unloaded module:", word)
				}
			}
		case ".info":
			for _, word := range chunks[4:] {
				word = strings.Trim(word, "\r\n")
				module, exists := conn.modules[word]
				if !exists {
					fmt.Fprintln(conn, "Module does not exist:", word)
				} else {
					fmt.Fprintf(conn, "%s: %s\n", word, module.Info)
				}
			}
		case ".loaded":
			fmt.Fprint(conn, "Loaded modules: ")
			for _, module := range conn.modules {
				if module.loaded {
					fmt.Fprint(conn, module.Name, " ")
				}
			}
			fmt.Fprint(conn, "\n")
		case ".list":
			fmt.Fprint(conn, "Modules: ")
			for _, module := range conn.modules {
				fmt.Fprint(conn, module.Name, " ")
			}
			fmt.Fprint(conn, "\n")
		}
	}

	//Add the callback to the connection
	c.AddCallback("privmsg", handler)
}
