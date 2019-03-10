package commands

import (
	"bufio"
	"bytes"
	"golang.org/x/crypto/ssh"
	"os"
	"strings"
	"time"
)

func AddConnection(host string, config *ssh.ClientConfig) (*ssh.Session, error) {

	// default port
	if !strings.Contains(host, ";") {
		host += ":22"
	}

	conn, err := ssh.Dial("tcp", host, config)

	if err != nil {
		return nil, err
	}

	return conn.NewSession()
}

//var (
//	clearLinux = func() {
//		cmd := exec.Command("clear") //Linux example, its tested
//		cmd.Stdout = os.Stdout
//		cmd.Run()
//	}
//
//	clearWindows = func() {
//		cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
//		cmd.Stdout = os.Stdout
//		cmd.Run()
//	}
//
//	clear = map[string]func(){
//		"linux", clearLinux,
//		clearWindows
//	}
//)

var (
	reader = bufio.NewReader(os.Stdin)
)

func StartTransmitting(connections []*ssh.Session) {
	ClearTerminal()

	for {
		rawBytes, err := reader.ReadBytes('\n')
		if err != nil {
			ClearTerminal()
			break
		}

		// ответ пишем сюда
		routineResults := make(chan string, 10)

		for i := 0; i < len(connections); i++ {
			go func(conn *ssh.Session) {
				routineResults <- runCmd(conn, string(rawBytes))
			}(connections[i])
		}

		// Время ождания команды
		timeout := time.After(5 * time.Second)
		result := make([]string, 10)
		select {
		case res := <-routineResults:
			result = append(result, res)
		case <-timeout:
			result = append(result, "Timed out!")
			return
		}

	}
}

func runCmd(session *ssh.Session, cmd string) string {
	var stdoutBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Run(cmd)

	return stdoutBuf.String()
}

func ClearTerminal() {
	// Обещают очистить терминал
	print("\033[H\033[2J")
}
