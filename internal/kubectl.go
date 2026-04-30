package internal

import (
	"fmt"
	"kod/types"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

func CreateSecretFromDockerJsonInHomeLocation(namespace string, secretName string) error {
	// TODO check if it already exists, and don't create it if so (just output an info note)

	// construct kubectl command
	homedir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting user's home directory")
		return err
	}
	kubectlExec := exec.Command("bash", "-c", "kubectl create secret docker-registry "+secretName+" --from-file="+filepath.Join(homedir, ".docker", "config.json")+" -n "+namespace)
	fmt.Println("Executing", kubectlExec.String())
	var kubectlOutput SaveOutput
	kubectlOutput.NoEchoToStdOut = true
	kubectlExec.Stdin = os.Stdin
	kubectlExec.Stdout = &kubectlOutput
	kubectlExec.Stderr = os.Stderr
	// Execute
	err = kubectlExec.Run()
	// Report any errors
	if err != nil {
		fmt.Println("Error executing kubectl create secret. Details:", kubectlOutput.String())
		return err
	}

	return nil
}

// WaitForNamespaceToExist
/* Waits for a K8s namespace to exist.
 * If kubectl fails, returns an error.
 * Returns true if exists at or before timeout.
 * Returns false if it didn't exist by timeout
 */
func WaitForNamespaceToExist(namespace string, period int, maxAttempts int) (bool, error) {
	// do an initial wait
	time.Sleep(time.Duration(period) * time.Second)

	// now try-and-wait until timeout (effectively initial + timeout) is exceeded
	for i := 0; i < maxAttempts; i++ {
		exists, err := NamespaceExists(namespace)
		if err != nil {
			return false, err
		}
		if exists {
			return true, nil
		}
		time.Sleep(time.Duration(period) * time.Second)
	}

	return false, nil
}

func NamespaceExists(namespace string) (bool, error) {
	kubectlExec := exec.Command("bash", "-c", "kubectl get ns -oyaml "+namespace)
	//fmt.Println("Executing", kubectlExec.String())
	var kubectlOutput SaveOutput
	kubectlOutput.NoEchoToStdOut = true
	kubectlExec.Stdin = os.Stdin
	kubectlExec.Stdout = &kubectlOutput
	kubectlExec.Stderr = os.Stderr
	// Execute
	err := kubectlExec.Run()
	// Report any errors
	if err != nil {
		fmt.Println("Error executing kubectl get ns. Details:", kubectlOutput.String())
		return false, err
	}

	// attempt to parse result as YAML
	newNs := types.K8sNamespace{}
	err = yaml.Unmarshal(kubectlOutput.SavedOutput, &newNs)
	if err != nil {
		fmt.Println("Error unmarshalling kubectl get ns -oyaml output. Details:", kubectlOutput.SavedOutput)
		return false, err
	}

	if newNs.Metadata.CreationTimestamp != "" && newNs.Status.Phase == "Active" {
		return true, nil
	}

	return false, nil
}
