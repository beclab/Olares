package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"bytetrade.io/web3os/installer/pkg/core/logger"
	"github.com/pkg/errors"
	"github.com/saintfish/chardet"
	utilexec "k8s.io/utils/exec"
)

type Charset string

const (
	DEFAULT Charset = "DEFAULT"
	GBK     Charset = "GBK"
	UTF8    Charset = "UTF8"
	UTF16   Charset = "UTF16"
)

type CommandExecute interface {
	Run() (string, error)
	Exec() (string, error)
}

type CommandExecutor struct {
	name        string
	prefix      string
	cmd         []string
	exitCode    int
	printOutput bool
	printLine   bool
}

type PowerShellCommandExecutor struct {
	Commands    []string
	PrintOutput bool
	PrintLine   bool
}

func (p *PowerShellCommandExecutor) Run() (string, error) {
	var cmd = &CommandExecutor{
		name:        "powershell",
		prefix:      "-Command",
		cmd:         p.Commands,
		printOutput: p.PrintOutput,
		printLine:   p.PrintLine,
	}

	return cmd.run()
}

type DefaultCommandExecutor struct {
	Commands    []string
	PrintOutput bool
	PrintLine   bool
}

func (d *DefaultCommandExecutor) Run() (string, error) {
	var cmd = &CommandExecutor{
		name:        "cmd",
		prefix:      "/C",
		cmd:         d.Commands,
		printOutput: d.PrintOutput,
		printLine:   d.PrintLine,
	}

	return cmd.run()
}

func (d *DefaultCommandExecutor) RunCmd(name string, charset Charset) (string, error) {
	var cmd = &CommandExecutor{
		name:        name,
		cmd:         d.Commands,
		printOutput: d.PrintOutput,
		printLine:   d.PrintLine,
	}

	return cmd.runcmd(charset)
}

func (d *DefaultCommandExecutor) Exec() (string, error) {
	var cmd = &CommandExecutor{
		name:        "cmd",
		prefix:      "/C",
		cmd:         d.Commands,
		printOutput: d.PrintOutput,
		printLine:   d.PrintLine,
	}

	return cmd.exec()
}

func NewCommandExecutor(name, prefix string, args []string, printOutput, printLine bool) *CommandExecutor {
	return &CommandExecutor{
		name:        name,
		prefix:      prefix,
		cmd:         args,
		printOutput: printOutput,
		printLine:   printLine,
	}
}

func (command *CommandExecutor) getCmd() string {
	return strings.Join(command.cmd, " ")
}

func (command *CommandExecutor) runcmd(charset Charset) (string, error) {
	var res string
	var exec = utilexec.New()

	output, err := exec.Command(command.name, command.cmd...).Output()
	if command.printOutput {
		logger.Infof("[exec] CMD: %s, output: %s, err: %v", fmt.Sprintf("%s %v", command.name, command.cmd), string(output), err)
	}

	detector := chardet.NewTextDetector()
	result, _ := detector.DetectBest(output)
	res, _ = CharsetConverts(result.Charset, output, charset)

	if err != nil {
		logger.Debugf("[exec] CMD: %s, CHARSET: %s, OUTPUT: %s, error: %v", fmt.Sprintf("%s %v", command.name, command.cmd), result.Charset, res, err)
		return res, err
	}

	logger.Debugf("[exec] CMD: %s, CHARSET: %s, OUTPUT: %s", fmt.Sprintf("%s %v", command.name, command.cmd), result.Charset, res)
	return res, nil
}

func (command *CommandExecutor) run() (string, error) {
	args := append([]string{command.prefix}, command.cmd...)
	c := exec.Command(command.name, args...)

	out, err := c.StdoutPipe()
	if err != nil {
		return "", err
	}

	c.Stderr = c.Stdout

	if err := c.Start(); err != nil {
		command.exitCode = -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			command.exitCode = exitErr.ExitCode()
		}
		return "", err
	}

	var outputBuffer bytes.Buffer
	r := bufio.NewReader(out)

	for {
		line, err := r.ReadString('\n')
		line = strings.TrimSpace(line) + "\r"
		if err != nil {
			if err.Error() != "EOF" {
				logger.Errorf("[exec] read error: %s", err)
			}

			if command.printLine && line != "" {
				fmt.Println(line)
			}
			outputBuffer.WriteString(line)
			break
		}

		if command.printLine && line != "" {
			fmt.Println(line)
		}
		outputBuffer.WriteString(line)
	}

	err = c.Wait()
	if err != nil {
		command.exitCode = -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			command.exitCode = exitErr.ExitCode()
		}
	}
	res := outputBuffer.String()

	if command.printOutput {
		fmt.Printf("[exec] CMD: %s, OUTPUT: \n%s\n", c.String(), res)
	}
	logger.Debugf("[exec] CMD: %s, OUTPUT: %s", c.String(), res)
	return res, errors.Wrapf(err, "Failed to exec command: %s \n%s", command.getCmd(), res)
}

func (command *CommandExecutor) exec() (string, error) {
	args := append([]string{command.prefix}, command.cmd...)
	c := exec.Command(command.name, args...)

	out, err := c.StdoutPipe()
	if err != nil {
		return "", err
	}

	_, pipeWriter, err := os.Pipe()
	defer pipeWriter.Close()

	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = c.Stdout

	if err := c.Start(); err != nil {
		command.exitCode = -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			command.exitCode = exitErr.ExitCode()
		}
		return "", err
	}

	var outputBuffer bytes.Buffer
	r := bufio.NewReader(out)

	for {
		line, err := r.ReadString('\n')
		line = strings.TrimSpace(line) + "\r"
		if err != nil {
			if err.Error() != "EOF" {
				logger.Errorf("[exec] read error: %s", err)
			}

			if line != "\r" {
				_, err = pipeWriter.Write([]byte(line))
				pipeWriter.Close()
				if err != nil {
					break
				}
			}

			if command.printLine && line != "" {
				fmt.Println(line)
			}
			outputBuffer.WriteString(line)
			break
		}

		if line != "\n" && !strings.Contains(line, "\r") {
			_, err = pipeWriter.Write([]byte(line))
			if err != nil {
				break
			}
		}

		if command.printLine && line != "" {
			fmt.Println(line)
		}
		outputBuffer.WriteString(line)
	}

	err = c.Wait()
	if err != nil {
		command.exitCode = -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			command.exitCode = exitErr.ExitCode()
		}
	}
	res := outputBuffer.String()

	if command.printOutput {
		fmt.Printf("[exec] CMD: %s, OUTPUT: \n%s\n", c.String(), res)
	}
	logger.Debugf("[exec] CMD: %s, OUTPUT: %s", c.String(), res)
	return res, errors.Wrapf(err, "Failed to exec command: %s \n%s", command.getCmd(), res)
}
