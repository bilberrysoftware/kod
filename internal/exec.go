/*
Copyright © 2026 Kod project Contributors
*/
package internal

import (
	"os"
	"strings"
)

type SaveOutput struct {
	SavedOutput    []byte
	NoEchoToStdOut bool // defaults to false
	Prefix         string
}

func (so *SaveOutput) Write(p []byte) (n int, err error) {
	so.SavedOutput = append(so.SavedOutput, p...)
	if !so.NoEchoToStdOut {
		if so.Prefix != "" {
			_, _ = os.Stdout.Write([]byte(so.Prefix))
			bytesWritten := len(p)
			strContent := string(p)
			strContent = strings.ReplaceAll(strContent, "\n", "\n"+so.Prefix)
			if strings.HasSuffix(strContent, so.Prefix) {
				strContent = strContent[:len(strContent)-len(so.Prefix)]
			}
			_, _ = os.Stdout.Write([]byte(strContent))
			return bytesWritten, nil
		}
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
