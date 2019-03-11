package reader

import (
	"bufio"
	"os"
	"strings"
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
func Read() string{
	line, err := reader.ReadString('\n')
	if err != nil{
		panic(err)
	}

	return line
}
