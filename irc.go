package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"
)

type message struct {
	channel string
	author  string
	content string
	raw     string
}

var (
	messages []message
	channels []string = []string{"#pixel"}
	server   string   = "45.79.137.210" // irc.foonetic.net: 216.93.243.37, irc.esper.net: 45.79.137.210
	port     string   = "6667"
	nick     string   = "Heraliado"
	master   string   = "iuyte"
	password string   = "2000numbers"
	loggedIn bool     = false
	conn     net.Conn
)

func main() {
	address := strings.Join([]string{server, port}, ":")
	conn, _ = net.Dial("tcp", address)
	defer conn.Close()

	fmt.Println("Connecting...")
	time.Sleep(time.Millisecond * 500)
	go login()
	go terminal()
	for {
		message, _ := readBuffer()
		go handle(message)
	}
}

func readBuffer() (string, error) {
	m, e := bufio.NewReader(conn).ReadString('\n')
	m = strings.Split(m, "\r\n")[0]
	return m, e
}

func extract(m string) (*message, error) {
	var (
		i                        int = 0
		author, channel, content []string
		ml                       []string = strings.Split(m, "")
		e                        error    = errors.New(m)
	)

	if len(m) < 15 {
		return nil, e
	}
	for i++; ml[i] != "!" && i < len(ml)-1; i++ {
		author = append(author, string(ml[i]))
	}
	if i+2 > len(m) {
		return nil, e
	}
	for i++; ml[i] != "@" && i < len(ml)-1; i++ {
	}
	if i+2 > len(m) {
		return nil, e
	}
	for i++; ml[i] != " " && i < len(ml)-1; i++ {
	}
	if i+2 > len(m) {
		return nil, e
	}
	for i++; strings.Contains(" PRIVMSG ", ml[i]) && i < len(ml)-1; i++ {
	}
	if i+2 > len(m) {
		return nil, e
	}
	for i++; ml[i] == "#" && i < len(ml)-1; i++ {
		channel = append(channel, "#")
	}
	for ; ml[i] != " " && i < len(ml); i++ {
		channel = append(channel, ml[i])
	}
	for i += 2; i < len(ml); i++ {
		content = append(content, ml[i])
	}

	var (
		ch string = strings.Join(channel, "")
		au string = strings.Join(author, "")
		co string = strings.Join(content, "")
	)

	msg := message{ch, au, co, m}
	messages = append(messages, msg)
	if len(ch) == 0 || len(au) == 0 || len(co) == 0 {
		return nil, errors.New(m)
	}
	return &messages[len(messages)-1], nil
}

func send(text ...string) {
	t := strings.Join(append(text, "\r\n"), " ")
	fmt.Fprintf(conn, t)
	fmt.Print(t)
}

func sendMessage(channel, text string) {
	send("PRIVMSG", channel, ":"+text)
}

func handle(raw string) {
	if strings.Contains(raw, "PING") && strings.Index(raw, "PING") == 0 {
		fmt.Println(raw)
		go send("PONG" + strings.Split(raw, "PING")[1])
		return
	}
	m, err := extract(raw)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		fmt.Println("#" + m.channel + " <" + m.author + "> " + m.content)
	}
	go sendMessage("#"+m.channel, m.content)
}

func login() {
	time.Sleep(time.Millisecond * 3000)
	go send("USER", nick, nick, nick, ":Echo")
	time.Sleep(time.Millisecond * 2000)
	go send("NICK", nick)
	time.Sleep(time.Millisecond * 2000)
	go send("PRIVMSG nickserv :IDENTIFY", password)
	time.Sleep(time.Millisecond * 2000)
	go send("USER", nick, nick, nick, ":Echo")
	time.Sleep(time.Millisecond * 2000)
	go send("NICK", nick)
	time.Sleep(time.Millisecond * 2000)
	go send("PRIVMSG nickserv :IDENTIFY", password)
	for _, c := range channels {
		time.Sleep(time.Millisecond * 2000)
		go send("JOIN", c)
	}
	for {
		for i := 0; i < len(messages); i++ {
			if strings.Contains(messages[i].raw, "+i") {
				loggedIn = true
				break
			}
		}
	}
}

func terminal() {
	t := bufio.NewReader(os.Stdin)
	for {
		text, _ := t.ReadString('\n')
		ct := strings.Split(text, " ")
		channel := ct[0]
		msg := strings.Join(ct[1:], " ")
		if strings.ToLower(channel) != "/join" {
			go sendMessage(channel, msg)
			continue
		}
		go send("JOIN", strings.Trim(msg, " "))
	}
}
