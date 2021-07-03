package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Command  string
	Programs []Program
}
type Program struct {
	Id   string
	Week time.Weekday
	Time string
}
type Option struct {
	Time string
	Id   string
}

func initConfig() *Config {
	c := Config{}
	toml.DecodeFile("./config.toml", &c)
	return &c
}

func main() {
	config := initConfig()
	prog := make(chan Option)
	done := make(chan struct{})
	go func() {
		for _, v := range config.Programs {
			sendTime(v, prog)
		}
		close(done)
	}()
Wait:
	for {
		select {
		case opt := <-prog:
			out, err := exec.Command(config.Command, "rec", opt.Id, opt.Time).Output()
			if err != nil {
				fmt.Println(opt)
				fmt.Println(string(out))
				return
			}
			fmt.Println(string(out))
		case <-done:
			break Wait
		default:
		}
	}
}
func sendTime(prog Program, programTime chan<- Option) {
	today := time.Now()
	h, m, err := parseTime(prog.Time)
	if err == nil {
		for i := -7; i < 0; i++ {
			checkDay := today.AddDate(0, 0, i)
			if checkDay.Weekday() == prog.Week {
				recDay := time.Date(checkDay.Year(), checkDay.Month(), checkDay.Day(), h, m, 0, 0, time.Local)
				programTime <- Option{
					Id:   fmt.Sprintf("-id=%s", prog.Id),
					Time: fmt.Sprintf("-s=%s", recDay.Format("20060102150405")),
				}
				fmt.Println(recDay)
			}
		}
	} else {
		fmt.Println(err)
	}
}

func parseTime(timeStr string) (int, int, error) {
	timeSplit := strings.Split(timeStr, ":")
	if len(timeSplit) != 2 {
		return 0, 0, errors.New("time error:" + timeStr)
	}
	timeInt := make([]int, 2)

	for i, v := range timeSplit {
		intVal, e := strconv.Atoi(v)
		if e != nil {
			return 0, 0, errors.New("hour error:" + v)
		}
		timeInt[i] = intVal
	}
	return timeInt[0], timeInt[1], nil
}
