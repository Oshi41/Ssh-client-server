package commands

import (
	"../reader"
	"bytes"
	"golang.org/x/crypto/ssh"
	"strings"
	"time"
	"fmt"
	"strconv"
)

// Создаем ssh-соединение
func AddConnection(host string, config *ssh.ClientConfig) (*ssh.Client, error) {

	// default port - 22
	if !strings.Contains(host, ";") {
		host += ":22"
	}

	// Используем всегда tcp протокол
	return ssh.Dial("tcp", host, config)
}

func StartTransmitting(connections []*ssh.Client) {
	ClearTerminal()

	println("Press Ctrl + C to exit from transmitting mode")

	for {
		// Считали строку
		instructions := reader.Read()

		// ответ пишем сюда
		routineResults := make(chan cmdResults, 10)

		for i := 0; i < len(connections); i++ {
			go func(conn *ssh.Client) {
				routineResults <- runCmd(conn, instructions)

			}(connections[i])
		}

		// Время ождания команды
		timeout := time.After(5 * time.Second)
		result :=  make([] cmdResults, 0)

		// Ожидаем ответа от сервера
		select {
		case res := <-routineResults:
			result = append(result, res)
		case <-timeout:
			result = append(result, cmdResults{output:"Time out"})
			return
		}

		// тут будут сгруппированные резльтаты
		mappedResult := make(map[string][]string)

		for  _, item := range result  {

			// Новый текст, дописали результат
			if _, ok := mappedResult[item.output]; ok{
				mappedResult[item.output] = append(mappedResult[item.output], item.addr)
			} else {
				// Такой текст с результатом уже есть, дополняем
			mappedResult[item.output] = append(make([]string, 0) ,item.addr)
			}
		}

		// Ответ для одиночного соединения
		if len(connections) < 2{
			for key, _:= range mappedResult{
				fmt.Println(key)
			}
		} else {
			// тут пишем все ответы от всех серверов
			fmt.Println("responses:")

			// Прохлдим по всем значениям в мапе
			for key, value := range mappedResult{
				fmt.Println(key)
				fmt.Println("For " + strconv.Itoa(len(value)) + " server(s)")
			}
		}
	}
}


func runCmd(client *ssh.Client, instructions string) cmdResults {
	// Записали клиента
	var result cmdResults
	result.addr = client.RemoteAddr().String()

	// Открыли сессию
	session, err := client.NewSession()
	if err != nil {
		result.output = "Can't open session: " + err.Error()
		return result
	}

	// Сюда пишем ответы
	var out bytes.Buffer
	session.Stdout = &out
	session.Stderr = &out

	err = session.Run(instructions)

	// Записали результат
	result.output = out.String()

	if err != nil{
		result.output += " " + err.Error()
	}

	return result

}

//
type cmdResults struct {
	output string
	addr string
}

func ClearTerminal() {
	// Обещают очистить терминал
	print("\033[H\033[2J")
}
