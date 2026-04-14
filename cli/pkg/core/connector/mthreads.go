package connector

import "fmt"

func isMthreadsAIBook(execRuntime Runtime) bool {
	dmidecode, err := execRuntime.GetRunner().SudoCmd("command -v dmidecode", false, false)
	if err != nil {
		fmt.Printf("Error executing dmidecode command: %v\n", err)
		return false
	}

	if dmidecode == "" {
		fmt.Println("dmidecode command not found, cannot determine if it's an Mthreads AI Book.")
		return false
	}

	output, err := execRuntime.GetRunner().SudoCmd("dmidecode -s processor-manufacturer", false, false)
	if err != nil {
		fmt.Printf("Error executing dmidecode to get processor manufacturer: %v\n", err)
		return false
	}

	return output == "AIBOOK"
}

func isMthreadsAIBookM1000(execRuntime Runtime) bool {
	if !isMthreadsAIBook(execRuntime) {
		return false
	}

	output, err := execRuntime.GetRunner().SudoCmd("dmidecode -s processor-version", false, false)
	if err != nil {
		fmt.Printf("Error executing dmidecode to get processor version: %v\n", err)
		return false
	}

	return output == "M1000"
}
