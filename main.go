package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Command       string
	SkickaCommand string
	SavePath      string
	Programs      []Program
	Before        int
}
type Program struct {
	Id   string
	Week time.Weekday
	Time string
}
type Option struct {
	Time string
	Name string
	Id   string
}

func initConfig() *Config {
	c := Config{}
	toml.DecodeFile("./config.toml", &c)
	return &c
}

func main() {
	config := initConfig()
	fmt.Println(config)
	prog := make(chan Option)
	done := make(chan struct{})
	go func() {
		for _, v := range config.Programs {
			sendTime(v, prog, createBefore(config.Before))
		}
		close(done)
	}()
Wait:
	for {
		select {
		case opt := <-prog:
			if flg := recording(config.Command, opt); flg {
				uploadToCloud(config.SkickaCommand, config.SavePath)
			} else {
				break Wait
			}
		case <-done:
			break Wait
		default:
		}
	}
}
func createBefore(before int) int {
	if before == 0 {
		return 7
	}
	return before
}
func recording(command string, opt Option) bool {
	fmt.Println("recording start")
	out, err := exec.Command(command, "rec", "-output=mp3", opt.Id, opt.Time, opt.Name).Output()
	if err != nil {
		fmt.Println(fmt.Scanf("%s:%s", out, opt))
		return false
	}
	file, title := parseProgramName(string(out))
	renameTitle(file, title)
	fmt.Println(title)
	return true
}
func renameTitle(file, title string) {
	if len(file) == 0 || len(title) == 0 {
		fmt.Printf("lengh err:%s:%s\n", file, title)
	}
	d, f := filepath.Split(file)
	toPath := filepath.Join(d, title+"_"+f)
	os.Rename(file, toPath)
}
func parseProgramName(log string) (string, string) {
	title := ""
	file := ""
	for _, line := range strings.Split(log, "\n") {
		texts := strings.Split(line, "|")
		if len(texts) == 4 {
			if strings.Index(texts[2], "TITLE") == -1 {
				title = strings.Replace(texts[2], " ", "", -1)
			}
		} else if strings.Index(line, "mp3") != -1 {
			file = strings.Replace(line, " ", "", -1)
		}
	}
	return file, title
}
func uploadToCloud(sckicaCommand, savePath string) {
	fmt.Println("upload start")
	files, err := ioutil.ReadDir("./output")
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		fromPath := "output/" + file.Name()
		out, err := exec.Command(sckicaCommand, "upload", fromPath, savePath).Output()
		if err != nil {
			fmt.Println(fmt.Sprintf("%s:%s", err, out))
		} else {
			fmt.Println("remove:" + fromPath)
			os.Remove(fromPath)
		}
	}
}

func sendTime(prog Program, programTime chan<- Option, before int) {
	today := time.Now()
	h, m, err := parseTime(prog.Time)
	if err == nil {
		for i := -before; i < 0; i++ {
			checkDay := today.AddDate(0, 0, -1)
			if checkDay.Weekday() == prog.Week {
				recDay := time.Date(checkDay.Year(), checkDay.Month(), checkDay.Day(), h, m, 0, 0, time.Local)
				opt := Option{
					Id:   fmt.Sprintf("-id=%s", prog.Id),
					Time: fmt.Sprintf("-s=%s", recDay.Format("20060102150405")),
				}
				programTime <- opt
				fmt.Println(opt)
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
