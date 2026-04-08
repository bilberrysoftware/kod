/*
Copyright © 2026 Kod project Contributors
*/
package internal

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

/*
 * Check if a command exists. Uses where.exe on Windows, and which elsewhere.
 */
func CommandExists(command string) bool {
	operatingSystem := runtime.GOOS
	windows := false
	whichCmd := "which"
	if operatingSystem == "windows" {
		windows = true
		whichCmd = "where.exe"
	}
	whichExec := exec.Command(whichCmd, command)
	whichOutput, err := whichExec.Output()
	if err != nil {
		fmt.Println("Error using which to find command", command, "err:", err, "details:", whichOutput)
		os.Exit(1)
	}
	lastLine := strings.TrimSpace(string(whichOutput))
	if windows {
		return strings.HasSuffix(lastLine, command+".exe")
	}
	return strings.HasSuffix(lastLine, command)
}
