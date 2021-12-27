package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"time"

	"./edgo"
	"./edgo/watch"
)

type filterFlag []string

func (i *filterFlag) String() string {
	return fmt.Sprint([]string(*i))
}

func (i *filterFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var (
	ErrInterrupted = errors.New("main: interrupted")
	filters        filterFlag
)

func waitForInterrupt(shutdown watch.Shutdown) {
	sigc := make(chan os.Signal)
	defer close(sigc)

	signal.Notify(sigc, os.Interrupt)
	defer signal.Stop(sigc)

	select {
	case <-sigc:
		shutdown.Kill(ErrInterrupted)
	case <-shutdown.Dying():
		/*noop*/
	}
}

// To set colors on the VPC controller, a command like this is used.
// .\VPC_LED_Control.exe 3344 80CB 01 00 ff 00

type Command struct {
	cmd  string
	args []string
}

var (
	ColorIndex = []string{"00", "40", "80", "ff"}

	White := Command{"C:\\Program Files (x86)\\VPC Software Suite\\tools\\VPC_LED_Control.exe", []string{"3344", "80CB", "01", ColorIndex[3], ColorIndex[3], ColorIndex[3]}},

	CmdMapping = map[string]Command{
		"Docked":      White,
		"Docked":      White,
		"Docked":      White,
		"FSDJump":      Command{"C:\\Program Files (x86)\\VPC Software Suite\\tools\\VPC_LED_Control.exe", []string{"3344", "80CB", "01", ColorIndex[1], ColorIndex[3], ColorIndex[1]}},
		"FuelScoop":    Command{"C:\\Program Files (x86)\\VPC Software Suite\\tools\\VPC_LED_Control.exe", []string{"3344", "80CB", "01", ColorIndex[2], ColorIndex[2], ColorIndex[0]}},
		"HeatDamage":   Command{"C:\\Program Files (x86)\\VPC Software Suite\\tools\\VPC_LED_Control.exe", []string{"3344", "80CB", "01", ColorIndex[3], ColorIndex[1], ColorIndex[0]}},
		"HeatWarning":  Command{"C:\\Program Files (x86)\\VPC Software Suite\\tools\\VPC_LED_Control.exe", []string{"3344", "80CB", "01", ColorIndex[2], ColorIndex[1], ColorIndex[1]}},
		"HullDamage":   Command{"C:\\Program Files (x86)\\VPC Software Suite\\tools\\VPC_LED_Control.exe", []string{"3344", "80CB", "01", ColorIndex[3], ColorIndex[0], ColorIndex[0]}},
		"Interdicted":  Command{"C:\\Program Files (x86)\\VPC Software Suite\\tools\\VPC_LED_Control.exe", []string{"3344", "80CB", "01", ColorIndex[3], ColorIndex[1], ColorIndex[1]}},
		"Interdiction": Command{"C:\\Program Files (x86)\\VPC Software Suite\\tools\\VPC_LED_Control.exe", []string{"3344", "80CB", "01", ColorIndex[2], ColorIndex[1], ColorIndex[1]}},
		"UnderAttack":  Command{"C:\\Program Files (x86)\\VPC Software Suite\\tools\\VPC_LED_Control.exe", []string{"3344", "80CB", "01", ColorIndex[3], ColorIndex[1], ColorIndex[1]}},
	}
)

// .\VPC_LED_Control.exe 3344 80CB 01 00 ff 00
func ChangeVPColor(events chan Command, shutdown watch.Shutdown) {
	for {
		select {
		case e := <-events:
			cmd := exec.Command(e.cmd, e.args...)
			if err := cmd.Run(); err != nil {
				log.Println("Error: ", err)
			}

		case <-shutdown.Dying():
			return
		}
	}
}

func HandleEvents(events chan interface{}, shutdown watch.Shutdown) {
	startTime := time.Now()
	cmd := make(chan Command)
	go ChangeVPColor(cmd, shutdown)

	for {
		select {
		case e := <-events:
			t := edgo.GetEventTimestamp(e)
			if t2, err := time.Parse(time.RFC3339, t); err == nil || t2.After(startTime) {
				name := edgo.GetEventName(e)
				log.Println(name, ":", e)
				if v, ok := CmdMapping[name]; ok {
					cmd <- v
				}
			}

		case <-shutdown.Dying():
			close(cmd)
			return
		}
	}
}

func main() {
	flag.Var(&filters, "f", "Filtered events.")
	flag.Parse()

	var directory string
	if flag.NFlag() > 1 {
		directory = flag.Arg(1)
	} else if homedir, err := os.UserHomeDir(); err == nil {
		directory = filepath.Join(homedir, "Saved Games", "Frontier Developments", "Elite Dangerous")
	}
	if directory == "" {
		fmt.Printf("Usage: %s <path to elite dangerous journal>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}

	fmt.Printf("Using: %s %s\n", filepath.Base(os.Args[0]), directory)

	shutdown := watch.NewShutdown()
	w := edgo.NewEliteWatcher(directory, shutdown)
	defer w.Close()

	if len(filters) > 0 {
		// Only add event filters if they have been specified on the comand line.
		w.EventFilter = make(map[string]struct{})
		for _, v := range filters {
			w.EventFilter[v] = struct{}{}
			log.Println("main: filter ", v)
		}
		w.EventFilter["NavRoute"] = struct{}{}
		log.Println("main: filter NavRoute")
	}

	go w.Main()
	go HandleEvents(w.Journals, shutdown)

	waitForInterrupt(shutdown)
	log.Println("done...")
}

/*
go get github.com/fsnotify/fsnotify

GOOS=windows GOARCH=amd64 go build -ldflags "-H windowsgui"
*/
