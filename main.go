package main

import (
	"./commands"
	"./crypto"
	"./parser"
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
	"strings"
)

var (
	sessions = make([]*ssh.Session, 0)
)

func main() {
	// Печатаем инфу
	fmt.Println(parser.App.Help)

	for {
		command, err := parser.App.Parse(readline())

		writeAndShutdown(err)

		switch command {

		case parser.Exit.FullCommand():
			for i := 0; i < len(sessions); i++ {
				sessions[i].Close()
			}

			os.Exit(1)

		case parser.RecreateSsh.FullCommand():
			crypto.GenerateNew()
			break

		case parser.AddConn.FullCommand():
			var config *ssh.ClientConfig = nil

			if *parser.AddConnWithPass {
				config, err = crypto.GetPasswordConfig(*parser.AddConnName, *parser.AddConnPass)
			} else {
				config, err = crypto.GetSshConfig(*parser.AddConnName)
			}

			writeAndShutdown(err)
			conn, err := commands.AddConnection(*parser.AddConnHost, config)

			if err != nil || conn == nil {
				fmt.Println(err)
			}

			sessions = append(sessions, conn)
			fmt.Println("Connection established")

		case parser.StartTransmitting.FullCommand():

			break
		}
	}
}

var (
	reader = bufio.NewReader(os.Stdin)
)

func readline() []string {

	rawBytes, err := reader.ReadBytes('\n')

	writeAndShutdown(err)

	// убираем переносы строк
	splitted := strings.Split(strings.Replace(string(rawBytes), "\n", "", -1), " ")

	return splitted
}

func writeAndShutdown(err error) {
	// При ошибке просто закрываю приложение, впадлу обрабатывать
	if err != nil {
		fmt.Print(err)
		os.Exit(0)
	}
}

///////////////////////////////
//
///////////////////////////////
