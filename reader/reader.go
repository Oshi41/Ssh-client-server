package reader

import (
	"bufio"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

var (
	reader = bufio.NewReader(os.Stdin)
)

// Считывает строчку из консоли и разбивает по пробелам
func ReadParsed() []string {
	// убираем переносы строк
	return strings.Split(strings.Replace(Read(), "\n", "", -1), " ")
}

// Считывает строчку из консоли
func Read() string {
	line, err := reader.ReadString('\n')
	if err != nil {
		panic(err)
	}

	return line
}

// Читает сырые байты
func ReadBytes() []byte {
	bytes, err := reader.ReadBytes('\n')
	if err != nil {
		panic(err)
	}

	return bytes
}

// ожидает ввода CTRL + C
func IsEscaped() chan bool {
	result := make(chan bool)

	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		result <- true
	}()

	return result
}
