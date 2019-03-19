package commands

import (
	"../reader"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"os/exec"
	"log"
	"os/signal"
)

type connectedClient struct {
	client    *ssh.Client
	session   *ssh.Session
	stdin     io.WriteCloser
	stdout    io.Reader
	stdoutErr io.Reader
}

// Сюда передаю список адресов, куда буду транслировать команды
func StartTranslate(addresses [] string, config *ssh.ClientConfig)  {
	connections := make([] *ssh.Client, 0)
	for _, addr := range addresses {
		client, err := ssh.Dial("tcp", addr, config)
		if err != nil{
			log.Println(err)
			continue
		}

		connections = append(connections, client)
	}

	if len(connections) > 0{
		startTranslateInner(connections)
	} else {
		log.Println("No available connections")
		return
	}

	for _, con := range connections{
		con.Close()
	}
}

func startTranslateInner(clients []*ssh.Client) {
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
	escape := make(chan os.Signal, 1)
	signal.Notify(escape, os.Interrupt)

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
			log.Println("request for pseudo terminal failed: %s", err)
			continue
		}

		// Start remote shell
		if err := session.Shell(); err != nil {
			log.Println("failed to start shell: %s", err)
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
