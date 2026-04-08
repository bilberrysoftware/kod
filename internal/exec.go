/*
Copyright © 2026 Kod project Contributors
*/
package internal

import "os"

type SaveOutput struct {
	SavedOutput    []byte
	NoEchoToStdOut bool // defaults to false
}

func (so *SaveOutput) Write(p []byte) (n int, err error) {
	so.SavedOutput = append(so.SavedOutput, p...)
	if !so.NoEchoToStdOut {
		return os.Stdout.Write(p)
	}
	return len(p), nil
}

func (so *SaveOutput) String() string {
	return string(so.SavedOutput)
}

func (so *SaveOutput) Clear() {
	so.SavedOutput = []byte{}
}
