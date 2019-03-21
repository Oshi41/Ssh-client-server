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
	"sync"
)

var (
	// Список всех доступных подключений
	clients = make([]*ssh.Client, 0)

	// синхронизатор для добавления
	waitGroup sync.WaitGroup

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

		case parser.Exit.FullCommand():
			os.Exit(1)

		case parser.RecreateSsh.FullCommand():
			fmt.Println("Key generating started...")
			err := keys.GenerateNew()

			if err != nil {
				log.Println(err)
			} else {
				fmt.Println("New key was generated")
			}
			break

		case parser.CloseConn.FullCommand():
			toRemove := *parser.CloseConnHost
			// Удаляю как по имени, так и по адресу
			for _, value := range clients {
				if strings.Contains(value.RemoteAddr().String(), toRemove) {
					closeConn(value)
				}

				if strings.Contains(value.User(), toRemove) {
					closeConn(value)
				}
			}

			break

		case parser.AddConn.FullCommand():

			go func() {
				client, err := commands.GetClient(*parser.AddConnHost,
					*parser.AddConnName,
					*parser.AddConnPass,
					*parser.AddConnWithPass,
					&waitGroup)

				if err != nil {
					log.Println(err)
				} else {
					clients = append(clients, client)
					log.Println("Successfully connected ", client.User())
				}
			}()

			break

		case parser.StartTransmitting.FullCommand():

			// Todo ask if we want to await for connections

			if len(clients) > 0 {
				commands.StartTranslate(clients)
			}

			break

		////////////////////////
		// ONLY FOR DEBUG
		////////////////////////
		case parser.Debug.FullCommand():
			//config, _ := keys.GetPasswordConfig("iu8_82_14", "1qazXSW@")
			//sessions = append(sessions, "185.20.227.83:2222")
			//commands.StartTranslate(sessions, config)
		}
	}
}

func closeConn(client *ssh.Client) {
	err := client.Close()
	if err != nil {
		log.Println(err)
	}

	for i := 0; i < len(clients); i++ {
		if clients[i] == client {
			clients = append(clients[:i], clients[i+1:]...)
			log.Println("Closed connection with ", client.User())
			break
		}
	}
}
