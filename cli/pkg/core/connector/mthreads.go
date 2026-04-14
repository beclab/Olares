package connector

import (
	"fmt"
	"os/exec"
)

func isMthreadsAIBook(cmdExec func(s string) (string, error)) bool {
	dmidecode, err := cmdExec("command -v dmidecode")
	if err != nil {
		fmt.Printf("Error executing dmidecode command: %v\n", err)
		return false
	}

	if len(dmidecode) == 0 {
		fmt.Println("dmidecode command not found, cannot determine if it's an Mthreads AI Book.")
		return false
	}

	output, err := cmdExec("dmidecode -s processor-manufacturer")
	if err != nil {
		fmt.Printf("Error executing dmidecode to get processor manufacturer: %v\n", err)
		return false
	}

	return output == "AIBOOK"
}

func isMthreadsAIBookM1000(cmdExec func(s string) (string, error)) bool {
	if !isMthreadsAIBook(cmdExec) {
		return false
	}

	output, err := cmdExec("dmidecode -s processor-version")
	if err != nil {
		fmt.Printf("Error executing dmidecode to get processor version: %v\n", err)
		return false
	}

	return output == "M1000"
}

func IsMThreadsAIBookM1000Local() bool {
	return isMthreadsAIBookM1000(func(s string) (string, error) {
		out, err := exec.Command("sh", "-c", s).Output()
		if err != nil {
			return "", err
		}
		return string(out), nil
	})
}

func IsMThreadsAIBookM1000(execRuntime Runtime) bool {
	return isMthreadsAIBookM1000(func(s string) (string, error) {
		return execRuntime.GetRunner().SudoCmd(s, false, false)
	})
}
