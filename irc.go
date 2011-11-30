package irc

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

/*
package irc for creating IRC bots. Has a concept of modules and commands,
and conforms to the io.Writer protocol.
*/

//Callback functions
type Callback func(*Connection, []string)

//Struct for info about the irc connection
type Info struct {
	Channel string
	Nick    string
	AltNick string
	Server  string
}

//Struct for our irc connection
type Connection struct {
	Conn      *net.TCPConn
	Info      Info
	buf       []byte
	callbacks map[string][]Callback
	modules   map[string]*Module
	ticker    *time.Ticker
}

//Sends a nickname string
func (conn *Connection) SendNick(nick string) {
	fmt.Fprintln(conn.Conn, "nick", nick)
}

//Joins a channel
func (conn *Connection) JoinChannel(channel string) {
	fmt.Fprintln(conn.Conn, "join", channel)
}

//Writes the bytes into the buffer
func (conn *Connection) Write(p []byte) (n int, err os.Error) {
	//append bytes
	conn.buf = append(conn.buf, p...)

	//attempt to flush any messages sent
	conn.Flush()

	//bytes were always appended, so no errors writing
	return len(p), nil
}

//Flushes any fully formed messages out to the channel
func (conn *Connection) Flush() {
	//split the messages in the buffer into individual messages
	messages := bytes.SplitAfter(conn.buf, []byte{'\n'})

	//if theres only one message, then theres no newline yet
	//so continue to buffer
	if len(messages) == 1 {
		return
	}

	//chop off the last message because it's just a blank string
	for _, message := range messages[:len(messages)-1] {

		//attempt to send the message
		if n, err := conn.SendMessage(string(message)); err != nil {
			//if an error occurs, chop off the bit that was sent
			conn.buf = conn.buf[n:]
			return
		}
		//chop off the message from the buffer
		conn.buf = conn.buf[len(message):]
	}

	return
}

//Sends a fully formed message to the channel
func (conn *Connection) SendMessage(message string) (n int, err os.Error) {
	var i, c, sn int
	for i < len(message) {
		c = i + 200
		if c > len(message) {
			c = len(message)
		}
		n, err = conn.sendMessage(message[i:c])
		if err != nil {
			return
		}
		sn += n
		i = c
	}
	return sn, err
}

func (conn *Connection) sendMessage(message string) (n int, err os.Error) {
	//grab a tick
	<-conn.ticker.C

	//Prime the message. If there are any problems, we sent 0 bytes of the message
	if n, err = conn.prefixPrivmsgToChannel(); err != nil {
		return 0, err
	}

	//Send it down and check for errors
	n, err = fmt.Fprint(conn.Conn, message)

	//if the message didnt end with a newline, add that on now
	if !strings.HasSuffix(message, "\n") {
		fmt.Fprint(conn.Conn, "\n")
	}

	return
}

//Sends an emote to the channel
func (conn *Connection) Emote(message string) (n int, err os.Error) {
	//Prime the message. If there are any problems, we sent 0 bytes of the message
	if n, err = conn.prefixPrivmsgToChannel(); err != nil {
		return 0, err
	}

	//Prefix with action
	if n, err = fmt.Fprint(conn.Conn, "\u0001ACTION "); err != nil {
		return 0, err
	}

	//Send the message down the pipe
	n, err = fmt.Fprint(conn.Conn, strings.Trim(message, "\n\r"))

	//Terminate the message
	fmt.Fprint(conn.Conn, "\u0001\n")

	return
}

//Prefixes a message with the privmsg to the channel
func (conn *Connection) prefixPrivmsgToChannel() (n int, err os.Error) {
	return fmt.Fprint(conn.Conn, "privmsg ", conn.Info.Channel, " :")
}

//Send the login packet to the IRC server
func (conn *Connection) SendLogin() {
	conn.SendNick(conn.Info.Nick)
	fmt.Fprintln(conn.Conn, "user okco okco okco okco")
}

//Creates a new Connection object ready to go
func NewConnection(info Info) (conn *Connection, err os.Error) {
	//Resolve the address of the irc server
	addr, err := net.ResolveTCPAddr("tcp", info.Server)
	if err != nil {
		return nil, err
	}

	//Connect to the server
	tcpConn, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		return nil, err
	}

	//Create the new struct for the connection
	conn = &Connection{
		Conn:      tcpConn,
		Info:      info,
		buf:       make([]byte, 0),
		callbacks: make(map[string][]Callback),
		modules:   make(map[string]*Module),
		ticker:    time.NewTicker(1e9),
	}

	//Send login packets
	conn.SendLogin()

	//Set up auto join of the channel, and alternate nickanmes
	conn.SetUpAutoJoin()
	conn.SetUpAltNick()

	//Return our new struct
	return
}

//Convenience method for setting up join on connect
func (conn *Connection) SetUpAutoJoin() {
	tmp := func(c *Connection, s []string) {
		c.JoinChannel(conn.Info.Channel)
	}
	for _, v := range []string{"376", "422"} {
		conn.AddCallback(v, tmp)
	}
}

//Convenience method for setting up sending alternate nickname
func (conn *Connection) SetUpAltNick() {
	conn.AddCallback("433", func(c *Connection, s []string) {
		c.SendNick(conn.Info.AltNick)
	})
}

//Grabs lines from the Connection and passes them to handlers
//Handles PING automatically
func (conn *Connection) Handle() {
	bufReader := bufio.NewReader(conn.Conn)
	for {
		cmd, err := bufReader.ReadString('\n')
		if err != nil {
			log.Print(conn, err)
			return
		}

		//Handle some basic commands
		chunks := strings.Split(cmd, " ")

		//Check for PING command
		if strings.HasPrefix(cmd, "PING") {
			fmt.Fprintln(conn.Conn, "PONG", chunks[1])
			continue
		}

		//Break out if it only sent one word
		if len(chunks) == 1 {
			continue
		}

		if len(chunks) > 3 && len(chunks[3]) > 1 {
			chunks[3] = chunks[3][1:] //strip off the left :
		}
		chunks[len(chunks)-1] = strings.TrimSpace(chunks[len(chunks)-1])

		//Get the callbacks
		callbacks, exists := conn.callbacks[strings.ToLower(chunks[1])]
		if !exists {
			continue
		}

		//Handle the other commands
		for _, call := range callbacks {
			call(conn, chunks)
		}
	}
}

//Grabs the username from a full host specification
func GetUsername(addr string) string {
	return strings.TrimLeft(strings.Split(addr, "!")[0], ":")
}
