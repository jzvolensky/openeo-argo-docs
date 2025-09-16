/*
Automatic tool installation for openeo-argoworkflows by EODC

This Go program automates the installation of various tools required for
working with OpenEO and Argo Workflows. It checks for the presence of
required binaries, downloads them if they are not found, and sets them
up for use.

It also checks if they are installed and prompts if outdated.

Juraj Zvolenský
Eurac Research
*/
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	green  = "\033[32m"
	red    = "\033[31m"
	yellow = "\033[33m"
	reset  = "\033[0m"
)

func run(cmd string, args ...string) {
	c := exec.Command(cmd, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		fmt.Printf("%s❌ Command failed:%s %s %v\n", red, reset, cmd, args)
		os.Exit(1)
	}
}

func capture(cmd string, args ...string) string {
	c := exec.Command(cmd, args...)
	out, err := c.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func exists(bin string) bool {
	_, err := exec.LookPath(bin)
	return err == nil
}

func spinner(text string, done chan bool) {
	frames := []string{"-", "\\", "|", "/"}
	i := 0
	for {
		select {
		case <-done:
			fmt.Printf("\r%s✔%s %s\n", green, reset, text)
			return
		default:
			fmt.Printf("\r%s %s", frames[i], text)
			time.Sleep(100 * time.Millisecond)
			i = (i + 1) % len(frames)
		}
	}
}

func downloadFile(url, output string) {
	run("sh", "-c", fmt.Sprintf("curl -fL %s -o %s", url, output))
	info, err := os.Stat(output)
	if err != nil || info.Size() == 0 {
		fmt.Printf("%s❌ Download failed for %s%s\n", red, reset, output)
		os.Exit(1)
	}
}

func validateBinary(path string) bool {
	output := capture("file", path)
	if strings.Contains(output, "ELF") {
		return true
	}
	return false
}

func promptUpgrade(tool, currentVersion, targetVersion string) bool {
	currentVersion = strings.TrimSpace(currentVersion)
	targetVersion = strings.TrimSpace(targetVersion)

	if strings.Contains(currentVersion, targetVersion) {
		fmt.Printf("%s✔%s %s is already at target version (%s), skipping installation.\n", green, reset, tool, targetVersion)
		return false
	}

	fmt.Printf("%s⚠%s %s is already installed (version %s). Target version: %s\n",
		yellow, reset, tool, currentVersion, targetVersion)
	fmt.Print("Do you want to overwrite and install the new version? [y/N]: ")
	var input string
	fmt.Scanln(&input)
	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes"
}

func printVersion(cmd string, args ...string) string {
	if !exists(cmd) {
		fmt.Printf("%s✖%s %s not found\n", red, reset, cmd)
		return ""
	}
	output := capture(cmd, args...)
	lines := strings.Split(output, "\n")
	firstLine := strings.TrimSpace(lines[0])
	fmt.Printf("%s✔%s %s: %s\n", green, reset, cmd, firstLine)
	return firstLine
}

func cleanup() {
	tempFiles := []string{
		"kubectl",
		"minikube-linux-amd64",
		"get_helm.sh",
		"argo",
		"argo.gz",
	}

	for _, f := range tempFiles {
		if _, err := os.Stat(f); err == nil {
			os.Remove(f)
			fmt.Printf("%sℹ%s Removed temporary file: %s\n", yellow, reset, f)
		}
	}
}

func systemInfo() {
	fmt.Println("🚀 Automatic installation of tools required for openeo-argoworkflows")
	fmt.Println("===================================")
	fmt.Print("Do you want to continue with the installation? [y/N]: ")
	var input string
	fmt.Scanln(&input)
	input = strings.ToLower(strings.TrimSpace(input))
	if input != "y" && input != "yes" {
		fmt.Println("Installation aborted. Bye :(")
		os.Exit(0)
	}

	kernel := capture("uname", "-sr")
	arch := capture("uname", "-m")

	distro := "Unknown Linux"
	file, err := os.Open("/etc/os-release")
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		info := map[string]string{}
		for scanner.Scan() {
			line := scanner.Text()
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				info[parts[0]] = strings.Trim(parts[1], `"`)
			}
		}
		if val, ok := info["PRETTY_NAME"]; ok {
			distro = val
		}
	}

	fmt.Printf("🖥️  Distro: %s\n", distro)
	fmt.Printf("🔧 Kernel: %s\n", kernel)
	fmt.Printf("📦 Arch:   %s\n\n", arch)
}

func installKubectl() {
	targetVersion := "v1.34.0"
	if exists("kubectl") {
		currentVersion := capture("kubectl", "version", "--client")
		validateBinary("/usr/local/bin/kubectl")
		if currentVersion != "" && !promptUpgrade("kubectl", currentVersion, targetVersion) {
			fmt.Printf("%s✔%s Skipping kubectl installation\n", green, reset)
			return
		}
	}

	done := make(chan bool)
	go spinner("Installing kubectl...", done)

	run("sudo", "rm", "-f", "/usr/local/bin/kubectl", "kubectl")
	url := fmt.Sprintf("https://dl.k8s.io/release/%s/bin/linux/amd64/kubectl", targetVersion)
	downloadFile(url, "kubectl")
	run("chmod", "+x", "kubectl")
	run("sudo", "install", "-o", "root", "-g", "root", "-m", "0755", "kubectl", "/usr/local/bin/kubectl")
	validateBinary("/usr/local/bin/kubectl")
	done <- true
}

func installMinikube() {
	targetVersion := "v1.32.0"
	if exists("minikube") {
		currentVersion := capture("minikube", "version")
		validateBinary("/usr/local/bin/minikube")
		if currentVersion != "" && !promptUpgrade("minikube", currentVersion, targetVersion) {
			fmt.Printf("%s✔%s Skipping minikube installation\n", green, reset)
			return
		}
	}

	done := make(chan bool)
	go spinner("Installing Minikube...", done)

	run("sudo", "rm", "-f", "/usr/local/bin/minikube", "minikube-linux-amd64")
	url := fmt.Sprintf("https://storage.googleapis.com/minikube/releases/%s/minikube-linux-amd64", targetVersion)
	downloadFile(url, "minikube-linux-amd64")
	run("chmod", "+x", "minikube-linux-amd64")
	run("sudo", "install", "minikube-linux-amd64", "/usr/local/bin/minikube")
	validateBinary("/usr/local/bin/minikube")
	done <- true
}

func installHelm() {
	targetVersion := "v3.14.1"
	if exists("helm") {
		currentVersion := capture("helm", "version", "--short")
		validateBinary("/usr/local/bin/helm")
		if currentVersion != "" && !promptUpgrade("helm", currentVersion, targetVersion) {
			fmt.Printf("%s✔%s Skipping helm installation\n", green, reset)
			return
		}
	}

	done := make(chan bool)
	go spinner("Installing Helm...", done)

	run("sudo", "rm", "-f", "/usr/local/bin/helm", "get_helm.sh")
	url := "https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3"
	downloadFile(url, "get_helm.sh")
	run("chmod", "+x", "get_helm.sh")
	run("./get_helm.sh")
	validateBinary("/usr/local/bin/helm")
	done <- true
}

func installArgoCLI() {
	targetVersion := "v3.7.1"
	if exists("argo") {
		currentVersion := capture("argo", "version", "--short")
		validateBinary("/usr/local/bin/argo")
		if currentVersion != "" && !promptUpgrade("argo", currentVersion, targetVersion) {
			fmt.Printf("%s✔%s Skipping Argo CLI installation\n", green, reset)
			return
		}
	}

	done := make(chan bool)
	go spinner("Installing Argo CLI...", done)

	run("sudo", "rm", "-f", "/usr/local/bin/argo", "argo.gz", "argo-linux-amd64")
	url := fmt.Sprintf("https://github.com/argoproj/argo-workflows/releases/download/%s/argo-linux-amd64.gz", targetVersion)
	downloadFile(url, "argo.gz")
	run("gunzip", "-f", "argo.gz")
	run("chmod", "+x", "argo")
	run("sudo", "mv", "argo", "/usr/local/bin/argo")
	validateBinary("/usr/local/bin/argo")

	done <- true
}

func main() {
	systemInfo()

	installKubectl()
	installHelm()
	installMinikube()
	installArgoCLI()

	fmt.Println("\nSummary (versions):")
	printVersion("kubectl", "version", "--client=true")
	printVersion("helm", "version", "--short")
	printVersion("minikube", "version")
	printVersion("argo", "version", "--short")

	cleanup()
	fmt.Println("\n🎉 All tools are ready to use!")
}
