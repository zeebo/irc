package irc

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

//Struct for info about the irc connection
type IRCInfo struct {
	Channel string
	Nick    string
	AltNick string
	Server  string
}

//Struct for our irc connection
type IRCConnection struct {
	Conn   *net.TCPConn
	RecvCh chan string
	Info   IRCInfo
	buf    []byte
}

//Sends a nickname string
func (conn *IRCConnection) SendNick(nick string) {
	fmt.Fprintln(conn.Conn, "nick", nick)
}

//Joins a channel
func (conn *IRCConnection) JoinChannel(channel string) {
	fmt.Fprintln(conn.Conn, "join", channel)
}

//Writes the bytes into the buffer
func (conn *IRCConnection) Write(p []byte) (n int, err os.Error) {
	//append bytes
	conn.buf = append(conn.buf, p...)

	//attempt to flush any messages sent
	conn.Flush()

	//bytes were always appended, so no errors writing
	return len(p), nil
}

//Flushes any fully formed messages out to the channel
func (conn *IRCConnection) Flush() {
	//split the messages in the buffer into individual messages
	messages := bytes.SplitAfter(conn.buf, []byte{'\n'}, -1)

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
func (conn *IRCConnection) SendMessage(message string) (n int, err os.Error) {
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
func (conn *IRCConnection) Emote(message string) (n int, err os.Error) {
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

func (conn *IRCConnection) prefixPrivmsgToChannel() (n int, err os.Error) {
	return fmt.Fprint(conn.Conn, "privmsg ", conn.Info.Channel, " :")
}

//Creates a new IRCConnection object ready to go
func NewConnection(info IRCInfo) (conn *IRCConnection, err os.Error) {
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
	conn = &IRCConnection{
		Conn:   tcpConn,
		RecvCh: make(chan string),
		Info:   info,
		buf:    make([]byte, 0),
	}

	//Start grabbing lines and sendding them down the channel
	go conn.grabLines()

	//Send the login packet to the IRC server
	conn.SendNick(info.Nick)
	fmt.Fprintln(tcpConn, "user okco okco okco okco")

	//Return our new struct with the message channel
	return
}

//Grabs lines from the IRCConnection and throws them down the RecvCh
//Only sends down lines it didn't handle already. Handles PING, motd finished
//and nickname in use messages.
func (conn *IRCConnection) grabLines() {
	bufReader := bufio.NewReader(conn.Conn)
	for {
		cmd, err := bufReader.ReadString('\n')
		if err != nil {
			log.Print(err)
		}

		//Handle some basic commands
		chunks := strings.Split(cmd, " ", -1)

		//Check for PING command
		if strings.HasPrefix(cmd, "PING") {
			fmt.Fprintln(conn.Conn, "PONG", chunks[1])
			continue
		}

		//Handle the other commands
		switch chunks[1] {
		case "376":
			conn.JoinChannel(conn.Info.Channel)
		case "433":
			conn.SendNick(conn.Info.AltNick)
		default:
			conn.RecvCh <- cmd
		}
	}
}
