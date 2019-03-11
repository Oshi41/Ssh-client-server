package parser

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"github.com/chzyer/readline"
)

var (
	App = kingpin.New("SSH-client console application", "Supported commands: [conn] [close] [ssh] [cmd] [exit]")

	AddConn         = App.Command("conn", "Adds a new connection to ssh-server")
	AddConnHost     = AddConn.Arg("Host", "Address to SSH-server").Required().String()
	AddConnName     = AddConn.Arg("Name", "Your login").Default("Ksusha").String()
	AddConnWithPass = AddConn.Flag("withPass", "Should we use password instead ssh keyes").Default("false").Bool()
	AddConnPass     = AddConn.Arg("Pass", "Password to establish connection").Default("").String()

	CloseConn     = App.Command("close", "Closes all existing connections")
	CloseConnHost = CloseConn.Arg("Host", "Address to SSH-server").Required().String()

	RecreateSsh = App.Command("ssh", "Recreates existing SSH key")

	StartTransmitting = App.Command("cmd",
		"Switches console to transmitting commands. Press Shift+C to escape")

	Exit = App.Command("exit", "Exit application")

	completer *readline.PrefixCompleter
)
