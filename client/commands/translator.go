package commands

import (
	"../reader"
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"
)

type connectedClient struct {
	client    *ssh.Client
	session   *ssh.Session
	stdin     io.WriteCloser
	stdout    io.Reader
	stdoutErr io.Reader
	close     func()
}

// Сюда передаю список адресов, куда буду транслировать команды
func StartTranslate(clients []*ssh.Client) {
	log.Println("Press Ctrl + C to exit from transmitting mode")

	connections := openSessions(clients)
	writeToOutput(connections)

	escape := make(chan os.Signal, 1)
	signal.Notify(escape, os.Interrupt)

	for {
		select {
		case <-escape:
			log.Println("CTRL + C detected, cancel receive mode")
			for _, val := range connections {
				val.close()
			}

		default:
			line := reader.Read()
			translate(connections, line)
		}
	}

}

// Открывает соединения у всех клиентов многопоточно
func openSessions(clients []*ssh.Client) []*connectedClient {

	waitGroup := sync.WaitGroup{}
	waitGroup.Add(len(clients))
	channels := make(chan *connectedClient, len(clients))

	for _, client := range clients {
		go func(client *ssh.Client, wg *sync.WaitGroup) {
			defer wg.Done()

			con, err := createConnection(client)
			if err != nil {
				log.Println(err)
			} else {
				channels <- con
			}
		}(client, &waitGroup)
	}

	waitGroup.Wait()
	close(channels)

	result := make([]*connectedClient, 0)

	for elem := range channels {
		result = append(result, elem)
	}

	return result
}

// Копирую вывод соединений в консоль
func writeToOutput(clients []*connectedClient) {

	copyOutput := func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			text := scanner.Text()
			fmt.Println(text)
		}
	}

	for _, client := range clients {
		//go io.Copy(os.Stdout, client.stdout)
		//go io.Copy(os.Stderr, client.stdoutErr)

		go copyOutput(client.stdoutErr)
		go copyOutput(client.stdout)
	}
}

// асинхронно передаю команды
func translate(clients []*connectedClient, line string) {
	timeNow := time.Now()

	translateAsync := func(in io.WriteCloser,
		line string,
		now *time.Time,
		flag chan int) {

		// Ожидаем получения сигнала от канала
		// Все потоки стартуют одновременно
		select {
		case <-flag:
			io.WriteString(in, line)
			log.Println(time.Since(*now).Nanoseconds()/1000, "mks")
		}
	}

	flag := make(chan int)

	for _, client := range clients {
		go translateAsync(client.stdin, line, &timeNow, flag)
	}

	// При закрытии потока все должны выолняться одновременно
	close(flag)
}

// Создаю подклчюение и заполняю данными
func createConnection(client *ssh.Client) (*connectedClient, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	stdin, err := session.StdinPipe()
	if err != nil {
		return nil, err
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stdoutErr, err := session.StderrPipe()
	if err != nil {
		return nil, err
	}

	// Set up terminal modes
	// https://net-ssh.github.io/net-ssh/classes/Net/SSH/Connection/Term.html
	// https://www.ietf.org/rfc/rfc4254.txt
	// https://godoc.org/golang.org/x/crypto/ssh
	// THIS IS THE TITLE
	// https://pythonhosted.org/ANSIColors-balises/ANSIColors.html
	modes := ssh.TerminalModes{
		ssh.ECHO:  0, // Disable echoing
		ssh.IGNCR: 1, // Ignore CR on input.
	}

	if err := session.RequestPty("vt100", 80, 40, modes); err != nil {
		return nil, err
	}

	// Start remote shell
	if err := session.Shell(); err != nil {
		return nil, err
	}

	// connections = append(connections, &con)
	return &connectedClient{
		client:    client,
		session:   session,
		stdin:     stdin,
		stdout:    stdout,
		stdoutErr: stdoutErr,
		close: func() {
			session.Close()
		},
	}, nil
}
