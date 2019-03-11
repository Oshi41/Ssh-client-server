package main

import (
	"./commands"
	"./keys"
	"./parser"
	"./reader"
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
)

var (
	sessions = make([]*ssh.Client, 0)
)

func main() {
	// Печатаем инфу
	fmt.Println(parser.App.Help)

	for {
		command, err := parser.App.Parse(reader.ReadParsed())
		if err != nil {
			panic(err)
		}

		switch command {

		case parser.Exit.FullCommand():
			for i := 0; i < len(sessions); i++ {
				sessions[i].Close()
			}

			os.Exit(1)

		case parser.RecreateSsh.FullCommand():
			keys.GenerateNew()
			break

		case parser.AddConn.FullCommand():
			var config *ssh.ClientConfig

			if *parser.AddConnWithPass {
				config, err = keys.GetPasswordConfig(*parser.AddConnName, *parser.AddConnPass)
			} else {
				config, err = keys.GetSshConfig(*parser.AddConnName)
			}

			if err != nil {
				panic(err)
			}

			conn, err := commands.AddConnection(*parser.AddConnHost, config)

			if err != nil || conn == nil {
				fmt.Println(err)
			}

			sessions = append(sessions, conn)
			fmt.Println("Connection established")

		case parser.StartTransmitting.FullCommand():
			commands.StartTransmitting(sessions)
			break
		}
	}
}

