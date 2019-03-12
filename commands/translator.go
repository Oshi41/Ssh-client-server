package commands

import (
	"../reader"
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"strconv"
)

type cmdResponse struct {
	client   *ssh.Client
	response string
}

func StartTranslate(clients []*ssh.Client) {
	clearTerminal()

	println("Press Ctrl + C to exit from transmitting mode")

	sessions, stdins, stdouts := createAndMapConnections(clients)

	// Копирую вывод в консоль
	// TODO было бы круто сделать distinct вывод соощбений,а не вываливать все подряд
	for _, stdout := range stdouts {
		go io.Copy(os.Stdout, stdout)
	}

	// заводим выход из нашего режима
	escape := reader.IsEscaped()

	for {
		select {
		// Если натыкаемся на escape-последовательность - выхожим
		case <-escape:
			break

		// Иначе продолжаем работу в обычном режиме
		default:

			// Считали строку
			instructions := reader.ReadBytes()

			for _, value := range stdins {
				_, err := fmt.Fprintf(value, "%s", instructions)
				if err != nil {
					fmt.Print(err)
				}
			}
		}
	}

	// В конце закрыли соедиения
	for _, value := range sessions {
		value.Close()
	}
}

// Возвращает мапу с возможными открытыми ессиями
func createAndMapConnections(connections []*ssh.Client) (
	map[*ssh.Client]*ssh.Session,
	map[*ssh.Client]io.WriteCloser,
	map[*ssh.Client]io.Reader) {

	sessions := make(map[*ssh.Client]*ssh.Session)
	writers := make(map[*ssh.Client]io.WriteCloser)
	readers := make(map[*ssh.Client]io.Reader)

	for _, item := range connections {
		session, err := item.NewSession()

		if err != nil {
			fmt.Println("Can't open session with " + item.RemoteAddr().String())
		} else {
			writer, err := session.StdinPipe()
			reader, err := session.StdoutPipe()
			if err != nil {
				fmt.Println("Can't open session input pipe " + item.RemoteAddr().String())
			} else {
				err = session.Shell()
				if err != nil {
					fmt.Println("Can't open shell here - " + item.RemoteAddr().String())
				} else {
					// Добавляем и сессию и инпут тут, друг без друга добавлять бесполезно
					writers[item] = writer
					readers[item] = reader
					sessions[item] = session
				}
			}
		}
	}

	return sessions, writers, readers
}

// Очищаем терминал от истории команд
func clearTerminal() {

}

// Выполняет инструкции в заданной сессии. Возвращает клиента для удобства маппирования
func runCmd(client *ssh.Client, session *ssh.Session, cmd string) cmdResponse {
	// Сюда пишем ответы
	var out bytes.Buffer
	var eOut bytes.Buffer

	session.Stdout = &out
	session.Stderr = &eOut

	err := session.Run(cmd)

	result := out.String() + " " + eOut.String()

	if err != nil {
		result += " " + err.Error()
	}

	return cmdResponse{client, result}
}

func writeMessage(responses []cmdResponse) string {
	// переложил по уникальным сообщениям
	distinctMessages := make(map[string][]*ssh.Client)

	for _, cmd := range responses {
		clients, ok := distinctMessages[cmd.response]

		if ok {
			clients = append(clients, cmd.client)
		} else {
			clients = append(make([]*ssh.Client, 0), cmd.client)
		}

		distinctMessages[cmd.response] = clients
	}

	result := ""

	if len(distinctMessages) == 1 {
		for key := range distinctMessages {
			result = key
		}
	} else {
		result += "Responses:\n"
		for key, value := range distinctMessages {
			result += key + "\n" + "On " + strconv.Itoa(len(value)) + " connections\n"
		}

	}

	return result

}
