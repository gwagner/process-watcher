package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"gopkg.in/yaml.v2"
)

var config *Config
var stopChan = make(chan os.Signal, 1)

func main() {
	var configFlag = flag.String("config", "commands.yaml", "Load the config with commands to watch and keep alive")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	// Load the config
	if err := LoadConfig(*configFlag); err != nil {
		log.Fatalf("Failed to load config: %s", err)
	}

	signal.Notify(stopChan, os.Interrupt)
	go func() {
		<-stopChan
		cancel()
	}()

	wg := sync.WaitGroup{}
	wg.Add(len(config.Commands))
	for _, command := range config.Commands {
		go func(command Command) {
			defer wg.Done()
			if err := CommandKeepalive(ctx, command); err != nil {
				log.Fatalf("Failed to keep alive command: %s", err)
			}
		}(command)
	}

	wg.Wait()
}

func CommandKeepalive(ctx context.Context, command Command) error {

	if command.SleepSeconds > 0 {
		log.Printf("Sleeping for %d seconds before starting command: %s", command.SleepSeconds, command.Name)
		time.Sleep(time.Duration(command.SleepSeconds) * time.Second)
	}

	log.Printf("Spinning up process watcher for: %s", command.Name)
	cmd, err := RunCommand(command)
	if err != nil {
		return fmt.Errorf("error trying to run process %s: %s", command.Name, err.Error())
	}

	waitChan := make(chan error)
	go func() {
		for {
			waitChan <- cmd.Wait()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
				return err
			}
			return nil

		case <-waitChan:
			if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
				log.Printf("Trying to rerun process: %s\n", command.Name)

				if command.RetrySec > 0 {
					log.Printf("Sleeping for %d seconds before retry", command.RetrySec)
					time.Sleep(time.Duration(command.RetrySec) * time.Second)
				}

				cmd, err = RunCommand(command)
				if err != nil {
					return fmt.Errorf("error trying to rerun process %s: %s", command.Name, err.Error())
				}
			}
		}
	}
}

func RunCommand(command Command) (*exec.Cmd, error) {
	cmd := exec.Command("/bin/sh", "-c", command.Cmd)
	if command.ShowLog {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	return cmd, cmd.Start()
}

// LoadConfig loads the config file
func LoadConfig(configPath string) error {
	// read the config file
	f, err := os.Open(configPath)
	if err != nil {
		return err
	}

	d, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(d, &config); err != nil {
		return err
	}

	for i, v := range config.Commands {
		if v.RetrySec == 0 {
			v.RetrySec = 10
		}

		config.Commands[i] = v
	}

	return nil
}
