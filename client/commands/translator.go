package commands

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"../reader"
	"io"
	"log"
	"sync"
	"os"
	"os/signal"
)

// Сюда передаю список адресов, куда буду транслировать команды
func StartTranslate(clients []*ssh.Client) {
	log.Println("Press Ctrl + C to exit from transmitting mode")

	connections := openSessions(clients)

	escape := make(chan os.Signal, 1)
	signal.Notify(escape, os.Interrupt)
	condition := sync.NewCond(&sync.Mutex{})

	for {
		select {
		case <-escape:
			log.Println("CTRL + C detected, cancel receive mode")
			for _, val := range connections {
				val.close()
			}

		default:
			line := reader.ReadBytes()
			for _, client := range connections{
				go func(client * connectedClient, line []byte) {
					condition.L.Lock()
					condition.Wait()
				 	defer condition.L.Unlock()

					client.stdin.Write(line)
				}(client, line)
			}

			condition.Broadcast()
		}
	}

	//closeMap := make(map[*ssh.Client]func())
	//
	//for _, client := range clients {
	//	go func(item *ssh.Client) {
	//		session, err := item.NewSession()
	//		if err != nil {
	//			fmt.Println("Can't start session ", item.User())
	//			return
	//		}
	//		stdin, err := session.StdinPipe()
	//		if err != nil {
	//			fmt.Println("Can't get standart input ", item.User())
	//			return
	//		}
	//
	//		stdout, err := session.StdoutPipe()
	//		if err != nil {
	//			fmt.Println("Can't get standart output ", item.User())
	//			return
	//		}
	//
	//		stdoutErr, err := session.StderrPipe()
	//		if err != nil {
	//			fmt.Println("Can't get error output ", item.User())
	//			return
	//		}
	//
	//		// Set up terminal modes
	//		// https://net-ssh.github.io/net-ssh/classes/Net/SSH/Connection/Term.html
	//		// https://www.ietf.org/rfc/rfc4254.txt
	//		// https://godoc.org/golang.org/x/crypto/ssh
	//		// THIS IS THE TITLE
	//		// https://pythonhosted.org/ANSIColors-balises/ANSIColors.html
	//		modes := ssh.TerminalModes{
	//			ssh.ECHO:  0, // Disable echoing
	//			ssh.IGNCR: 1, // Ignore CR on input.
	//		}
	//
	//		if err := session.RequestPty("vt100", 80, 40, modes); err != nil {
	//			log.Println("request for pseudo terminal failed: %s", err)
	//			return
	//		}
	//
	//		// Start remote shell
	//		if err := session.Shell(); err != nil {
	//			log.Println("failed to start shell: %s", err)
	//			return
	//		}
	//
	//		log.Println("Created connection on ", item.User())
	//
	//		closeMap[item] = func() {
	//			err = session.Close()
	//			if err != nil {
	//				log.Println(err)
	//			}
	//
	//			err = stdin.Close()
	//			if err != nil {
	//				log.Println(err)
	//			}
	//		}
	//
	//		go func() {
	//			//_, err = io.Copy(os.Stdout, stdout)
	//			//if err != nil{
	//			//	log.Println("ERROR", err)
	//			//}
	//			for {
	//				buffer := make([]byte, 32*1024)
	//				_, err := stdout.Read(buffer)
	//				if err != nil {
	//					log.Println(err)
	//					return
	//				}
	//
	//				os.Stdout.Write(buffer)
	//				//_, err = fmt.Fprintln(os.Stdout, item.User(), "\n", string(buffer), "\n")
	//				//if err != nil{
	//				//	log.Println(err)
	//				//	return
	//				//}
	//			}
	//		}()
	//
	//		go func() {
	//			_, err = io.Copy(os.Stderr, stdoutErr)
	//			if err != nil {
	//				log.Println("ERROR", err)
	//			}
	//
	//		}()
	//
	//		go func() {
	//			for {
	//				select {
	//				case text := <-acceptChannel:
	//					stdin.Write(text)
	//				}
	//			}
	//		}()
	//
	//	}(client)
	//}
	//

}

type connectedClient struct {
	client    *ssh.Client
	session   *ssh.Session
	stdin     io.WriteCloser
	stdout    io.Reader
	stdoutErr io.Reader
	close 	  func()
}

func openSessions(clients []*ssh.Client) [] *connectedClient  {
	wg := sync.WaitGroup{}
	channels := make(chan *connectedClient, len(clients))

	for _, client := range clients {
		go func(client *ssh.Client) {
			wg.Add(1)
			defer wg.Done()

			session, err := client.NewSession()
			if err != nil {
				fmt.Println("Can't start session ", client.User())
				return
			}
			stdin, err := session.StdinPipe()
			if err != nil {
				fmt.Println("Can't get standart input ", client.User())
				return
			}

			stdout, err := session.StdoutPipe()
			if err != nil {
				fmt.Println("Can't get standart output ", client.User())
				return
			}

			stdoutErr, err := session.StderrPipe()
			if err != nil {
				fmt.Println("Can't get error output ", client.User())
				return
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
				return
			}

			// Start remote shell
			if err := session.Shell(); err != nil {
				log.Println("failed to start shell: %s", err)
				return
			}

			channels <- &connectedClient{
				client:client,
				session:session,
				stdin:stdin,
				stdout:stdout,
				stdoutErr:stdoutErr,
				close: func() {
					session.Close()
				},
			}

			log.Println("Opened shell on ", client.RemoteAddr())

		}(client)
	}

	wg.Wait()

	result := make([]* connectedClient, 0)

	for i := 0;i < len(channels) ;i++  {
		select {
			case con := <- channels:
				result = append(result, con)

		default:
			// looks like error here
		}
	}

	return result
}

//func startTranslateInner(clients []*ssh.Client) {
//	clearTerminal()
//	println("Press Ctrl + C to exit from transmitting mode")
//
//	connections := openSessions(clients)
//
//	// Копирую вывод в консоль
//	// TODO было бы круто сделать distinct вывод соощбений,а не вываливать все подряд
//	for _, session := range connections {
//		go io.Copy(os.Stdout, session.stdout)
//		go io.Copy(os.Stdout, session.stdoutErr)
//	}
//
//	// заводим выход из нашего режима
//	escape := make(chan os.Signal, 1)
//	signal.Notify(escape, os.Interrupt)
//
//	for {
//		select {
//		case <-escape:
//			// В конце закрыли соедиения
//			for _, value := range connections {
//				value.session.Close()
//			}
//			fmt.Println("Cancel translation mode")
//			return
//
//		default:
//			// считали строку
//			instructions := reader.ReadBytes()
//
//			// записали её в подсключения
//			for _, session := range connections {
//				_, err := fmt.Fprintf(session.stdin, "%s", instructions)
//
//				if err != nil {
//					fmt.Print(err)
//				}
//			}
//		}
//	}
//
//}
//
//func openSessions(connections []*ssh.Client) []connectedClient {
//	result := make([]connectedClient, 0)
//
//	for _, item := range connections {
//
//		addr := item.RemoteAddr().String()
//		session, err := item.NewSession()
//		if err != nil {
//			fmt.Println("Can't start session ", addr)
//			continue
//		}
//		stdin, err := session.StdinPipe()
//		if err != nil {
//			fmt.Println("Can't get standart input ", addr)
//			continue
//		}
//
//		stdout, err := session.StdoutPipe()
//		if err != nil {
//			fmt.Println("Can't get standart output ", addr)
//			continue
//		}
//
//		stdoutErr, err := session.StderrPipe()
//		if err != nil {
//			fmt.Println("Can't get error output ", addr)
//			continue
//		}
//
//		// Set up terminal modes
//		// https://net-ssh.github.io/net-ssh/classes/Net/SSH/Connection/Term.html
//		// https://www.ietf.org/rfc/rfc4254.txt
//		// https://godoc.org/golang.org/x/crypto/ssh
//		// THIS IS THE TITLE
//		// https://pythonhosted.org/ANSIColors-balises/ANSIColors.html
//		modes := ssh.TerminalModes{
//			ssh.ECHO:  0, // Disable echoing
//			ssh.IGNCR: 1, // Ignore CR on input.
//		}
//
//		if err := session.RequestPty("vt100", 80, 40, modes); err != nil {
//			log.Println("request for pseudo terminal failed: %s", err)
//			continue
//		}
//
//		// Start remote shell
//		if err := session.Shell(); err != nil {
//			log.Println("failed to start shell: %s", err)
//			continue
//		}
//
//		result = append(result, connectedClient{
//			session:   session,
//			client:    item,
//			stdin:     stdin,
//			stdout:    stdout,
//			stdoutErr: stdoutErr})
//	}
//
//	return result
//}
//
//// Очищаем терминал от истории команд
//func clearTerminal() {
//	cmd := exec.Command("cmd", "/c", "cls")
//	cmd.Stdout = os.Stdout
//	cmd.Run()
//}
