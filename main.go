package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	var user string
	var keyPath string
	var remoteHost string

	fs := flag.NewFlagSet("rdocker", flag.ExitOnError)
	fs.StringVar(&user, "u", "", "SSH user for the remote host")
	fs.StringVar(&keyPath, "k", "", "Path to SSH private key file (optional)")

	usage := func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <remote_host> -- <docker/docker-compose command>\n", os.Args[0])
		fs.PrintDefaults()
	}

	// コマンドライン引数がない場合は Usage のみを表示して終了
	if len(os.Args) == 1 {
		usage()
		os.Exit(0)
	}

	separatorIndex := -1
	for i, arg := range os.Args {
		if arg == "--" {
			separatorIndex = i
			break
		}
	}

	if separatorIndex == -1 {
		fmt.Println("Error: Separator '--' is required.")
		os.Exit(1)
	}

	err := fs.Parse(os.Args[1:separatorIndex])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	args := fs.Args()
	if len(args) == 0 {
		fmt.Println("Error: Remote host is required.")
		os.Exit(1)
	}
	remoteHost = args[len(args)-1]

	dockerCmd := strings.Join(os.Args[separatorIndex+1:], " ")
	if dockerCmd == "" {
		fmt.Println("Error: Docker command is required after '--'.")
		os.Exit(1)
	}

	if user == "" {
		fmt.Println("Error: SSH user (-u) is required.")
		os.Exit(1)
	}

	err = runDockerCommand(user, keyPath, remoteHost, dockerCmd)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func runDockerCommand(user, keyPath, remoteHost, dockerCmd string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}
	dirName := filepath.Base(currentDir)

	if err := validateCommand(dockerCmd); err != nil {
		return err
	}

	remoteDir, err := createRemoteTempDir(user, keyPath, remoteHost, dirName)
	if err != nil {
		return fmt.Errorf("failed to create remote directory: %v", err)
	}

	err = syncCurrentDirectory(user, keyPath, remoteHost, remoteDir)
	if err != nil {
		return fmt.Errorf("failed to sync current directory: %v", err)
	}

	output, err := executeRemoteDockerCmd(user, keyPath, remoteHost, remoteDir, dockerCmd)
	if err != nil {
		return fmt.Errorf("failed to execute Docker command: %v", err)
	}

	fmt.Println(output)
	return nil
}

func validateCommand(cmd string) error {
	if strings.HasPrefix(cmd, "docker-compose") {
		if !fileExists("docker-compose.yml") && !fileExists("compose.yaml") {
			return fmt.Errorf("docker-compose command can only be executed in a directory with a docker-compose.yml or compose.yaml file")
		}
	} else if strings.HasPrefix(cmd, "docker") {
		if !fileExists("Dockerfile") {
			return fmt.Errorf("docker command can only be executed in a directory with a Dockerfile")
		}
	}
	return nil
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

func createRemoteTempDir(user, keyPath, remoteHost, dirName string) (string, error) {
	remoteDir := fmt.Sprintf("/tmp/%s", dirName)
	sshArgs := buildSSHArgs(keyPath, fmt.Sprintf("%s@%s", user, remoteHost), fmt.Sprintf("mkdir -p %s", remoteDir))
	cmd := exec.Command("ssh", sshArgs...)

	fmt.Printf("Executing: ssh %s\n", strings.Join(sshArgs, " "))

	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return remoteDir, nil
}

func syncCurrentDirectory(user, keyPath, remoteHost, remoteDir string) error {
	rsyncArgs := []string{
		"-avz",
		"--delete",
		"--exclude", ".git",
		"./",
		fmt.Sprintf("%s@%s:%s", user, remoteHost, remoteDir),
	}

	if keyPath != "" {
		rsyncArgs = append([]string{"-e", fmt.Sprintf("ssh -i %s", keyPath)}, rsyncArgs...)
	}

	cmd := exec.Command("rsync", rsyncArgs...)

	fmt.Printf("Executing: rsync %s\n", strings.Join(rsyncArgs, " "))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func executeRemoteDockerCmd(user, keyPath, remoteHost, remoteDir, dockerCmd string) (string, error) {
	sshCmd := fmt.Sprintf("cd %s && sudo %s", remoteDir, dockerCmd)

	sshArgs := buildSSHArgs(keyPath, fmt.Sprintf("%s@%s", user, remoteHost), sshCmd)
	cmd := exec.Command("ssh", sshArgs...)

	fmt.Printf("Executing: ssh %s\n", strings.Join(sshArgs, " "))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run command: %v\nOutput: %s", err, string(output))
	}

	return string(output), nil
}

func buildSSHArgs(keyPath, destination, command string) []string {
	args := []string{}
	if keyPath != "" {
		args = append(args, "-i", keyPath)
	}
	args = append(args, destination, command)
	return args
}
