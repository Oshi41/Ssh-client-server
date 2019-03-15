package commands

import (
	"../reader"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"os/exec"
)

type connectedClient struct {
	client    *ssh.Client
	session   *ssh.Session
	stdin     io.WriteCloser
	stdout    io.Reader
	stdoutErr io.Reader
}

func StartTranslate(clients []*ssh.Client) {
	clearTerminal()
	println("Press Ctrl + C to exit from transmitting mode")

	connections := openSessions(clients)

	// Копирую вывод в консоль
	// TODO было бы круто сделать distinct вывод соощбений,а не вываливать все подряд
	for _, session := range connections {
		go io.Copy(os.Stdout, session.stdout)
		go io.Copy(os.Stdout, session.stdoutErr)
	}

	// заводим выход из нашего режима
	escape := reader.IsEscaped()

	for {
		select {
		case <-escape:
			// В конце закрыли соедиения
			for _, value := range connections {
				value.session.Close()
			}
			fmt.Println("Cancel translation mode")
			return

		default:
			// считали строку
			instructions := reader.ReadBytes()

			// записали её в подсключения
			for _, session := range connections {
				_, err := fmt.Fprintf(session.stdin, "%s", instructions)

				if err != nil {
					fmt.Print(err)
				}
			}
		}
	}

}

func openSessions(connections []*ssh.Client) []connectedClient {
	result := make([]connectedClient, 0)

	for _, item := range connections {

		addr := item.RemoteAddr().String()
		session, err := item.NewSession()
		if err != nil {
			fmt.Println("Can't start session ", addr)
			continue
		}
		stdin, err := session.StdinPipe()
		if err != nil {
			fmt.Println("Can't get standart input ", addr)
			continue
		}

		stdout, err := session.StdoutPipe()
		if err != nil {
			fmt.Println("Can't get standart output ", addr)
			continue
		}

		stdoutErr, err := session.StderrPipe()
		if err != nil {
			fmt.Println("Can't get error output ", addr)
			continue
		}

		err = session.Shell()
		if err != nil {
			fmt.Println("Can't start shell ", addr)
			continue
		}

		result = append(result, connectedClient{
			session:   session,
			client:    item,
			stdin:     stdin,
			stdout:    stdout,
			stdoutErr: stdoutErr})
	}

	return result
}

// Очищаем терминал от истории команд
func clearTerminal() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}
