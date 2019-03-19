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
	sessions = make([]string, 0)
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
			config, _ := keys.GetPasswordConfig("iu8_82_14", "1qazXSW@")
			sessions = append(sessions, "185.20.227.83:22")
			commands.StartTranslate(sessions, config)

		case parser.Exit.FullCommand():
			os.Exit(1)

		case parser.RecreateSsh.FullCommand():
			fmt.Println("Key generating started...")
			keys.GenerateNew()
			fmt.Println("New key was generated")
			break

		case parser.AddConn.FullCommand():
			addr := *parser.AddConnHost

			if !strings.Contains(addr, ":"){
				addr += ":22"
			}

			sessions = append(sessions, addr)

			log.Println("Total connections - ", len(sessions))

		case parser.StartTransmitting.FullCommand():
			var config *ssh.ClientConfig

			if *parser.AddConnWithPass {
				config, err = keys.GetPasswordConfig(*parser.AddConnName, *parser.AddConnPass)
			} else {
				config, err = keys.GetSshConfig(*parser.AddConnName)
			}

			if err != nil {
				log.Println("ERROR\n", err)
			} else {
				commands.StartTranslate(sessions, config)
			}

			break

		case parser.CloseConn.FullCommand():
			toRemove := *parser.CloseConnHost
			clientAddresses := make([]string, 0)
			deleted := false

			for index, client := range sessions {
				clientAddresses = append(clientAddresses, client)
				if strings.Contains(client, toRemove) {
					sessions = append(sessions[:index], sessions[index+1:]...)
					fmt.Println("Client was deleted - ", client)
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
