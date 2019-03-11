package commands

import (
	"../reader"
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"strconv"
	"time"
)

type cmdResponse struct {
	client   *ssh.Client
	response string
}

func StartTranslate(clients []*ssh.Client) {
	clearTerminal()

	println("Press Ctrl + C to exit from transmitting mode")

	for {
		// заводим выход из нашего режима
		escape := reader.IsEscaped()

		select {
		// Если натыкаемся на escape-последовательность - выхожим
		case <-escape:
			return

		// Иначе продолжаем работу в обычном режиме
		default:
			connections := createAndMapConnections(clients)
			// Считали строку
			instructions := reader.Read()

			buffer := make(chan cmdResponse, 0)

			// Запускаю выполнение в каждом потоке (routine)
			for key, value := range connections {
				go func(client *ssh.Client, session *ssh.Session, cmd string) {
					buffer <- runCmd(client, value, instructions)
				}(key, value, instructions)
			}

			// выставим таймаут (5 сек)
			timeout := time.After(5 * time.Second)
			results := make([]cmdResponse, 0)

			// Дождались ответа и записали результат
			select {
			case response := <-buffer:
				results = append(results, response)

			case <-timeout:
				results = append(results, cmdResponse{response: "Time out!"})
			}

			// Напечатали резльтат
			fmt.Println(writeMessage(results))

			for _, value := range connections {
				value.Close()
			}
		}
	}

}

// Возвращает мапу с возможными открытыми ессиями
func createAndMapConnections(connections []*ssh.Client) map[*ssh.Client]*ssh.Session {
	result := make(map[*ssh.Client]*ssh.Session)

	for _, item := range connections {
		session, err := item.NewSession()

		if err != nil {
			fmt.Println("Can't open session with " + item.RemoteAddr().String())
		} else {
			result[item] = session
		}
	}

	return result
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
