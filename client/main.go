package main

import (
	"./commands"
	"./keys"
	"./parser"
	"./reader"
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"strings"
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
			log.Print(err)
			continue
		}

		switch command {

		////////////////////////
		// ONLY FOR DEBUG
		////////////////////////
		case parser.Debug.FullCommand():
			config, _ := keys.GetPasswordConfig("iu8_82_14", "bar")
			conn, _ := commands.AddConnection("185.20.227.83:2222", config)
			sessions = append(sessions, conn)
			commands.StartTranslate(sessions)

		case parser.Exit.FullCommand():
			for i := 0; i < len(sessions); i++ {
				sessions[i].Close()
			}
			os.Exit(1)

		case parser.RecreateSsh.FullCommand():
			fmt.Println("Key generating started...")
			keys.GenerateNew()
			fmt.Println("New key was generated")
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
			if err != nil {
				fmt.Println(err)
			} else {
				sessions = append(sessions, conn)
				fmt.Println("Connection established")
			}

		case parser.StartTransmitting.FullCommand():
			commands.StartTranslate(sessions)
			break

		case parser.CloseConn.FullCommand():
			toRemove := *parser.CloseConnHost
			clientAddresses := make([]string, 0)
			deleted := false

			for index, client := range sessions {
				addr := client.RemoteAddr().String()
				clientAddresses = append(clientAddresses, addr)
				if strings.Contains(addr, toRemove) {
					client.Close()
					sessions = append(sessions[:index], sessions[index+1:]...)
					fmt.Println("Client was deleted")
					deleted = true
				}
			}

			if !deleted {
				fmt.Println("Can't find client ", toRemove, "in current connections:\n", clientAddresses)
			}

			break
		}
	}
}
