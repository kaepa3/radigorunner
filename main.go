package main

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Programs []Program
}
type Program struct {
	Id   string
	Week time.Weekday
	Time string
}

func initConfig() *Config {
	c := Config{}
	toml.DecodeFile("./config.toml", &c)
	return &c
}

func main() {
	config := initConfig()
	prog := make(chan string)
	done := make(chan struct{})
	go func() {
		for _, v := range config.Programs {
			fmt.Printf("%+v\n", v)
			sendTime(v, prog)
		}
		close(done)
	}()
Wait:
	for {
		select {
		case text := <-prog:
			if out, err := exec.Command(text).Output(); err != nil {
				fmt.Println(text)
				fmt.Println(out)
			}
		case <-done:
			break Wait
		default:
			break
		}
	}
}
func sendTime(prog Program, programTime chan<- string) {
	today := time.Now().AddDate(0, 0, -7)
	for i := -7; i < 0; i++ {
		checkDay := today.AddDate(0, 0, i)
		if checkDay.Weekday() == prog.Week {
			programTime <- checkDay.String()
		}
	}
}
