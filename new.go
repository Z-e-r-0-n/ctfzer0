// check whether the interface is there before implementing the intent files
// denforce.py doesnt do downlink via wired implement that too

package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const intent_dir string = "/etc/net/interfaces"
const global_conf string = "/etc/net/global.conf"

const systemd_net_dir string = "/etc/systemd/network"
const wpa_supplicant_dir string = "/etc/wpa_supplicant"
const hostapd_conf_dir = "/etc/hostapd"
const dnsmasq_conf_dir = "/etc/dnsmasq.d"

func write_file(path string, content string) error {
	err := os.WriteFile(path, []byte(content), 0666)
	if err != nil {
		return err
	}
	err = os.Chmod(path, 0666)
	if err != nil {
		return err
	}
	return nil
}

func run(command string) {
	c := exec.Command("sh", "-c", command)
	err := c.Run()
	if err != nil {
		fmt.Println("Command failed :", err)
	}
}

func clean_slate() {
	fmt.Println("[*] Cleaning old configurations...")

	files, _ := filepath.Glob(systemd_net_dir + "/*-enforced.network")

	for _, f := range files {
		err := os.Remove(f)
		if err != nil {
			fmt.Println("failed to remove :", f, err)
		}
	}

	run("systemctl stop hostapd")
	run("pkill hostapd")
	run("systemctl stop dnsmasq")

	hostapdFiles, _ := filepath.Glob(hostapd_conf_dir + "/*.conf")

	for _, f := range hostapdFiles {
		err := os.Remove(f)
		if err != nil {
			fmt.Println("failed to remove:", f, err)
		}
	}

	dnsmasqFiles, _ := filepath.Glob(dnsmasq_conf_dir + "/*conf")

	for _, f := range dnsmasqFiles {
		err := os.Remove(f)
		if err != nil {
			fmt.Println("failed to remove:", f, err)
		}
	}

	run("pkill wpa_supplicant")

}

func parse_intent(path string) map[string]string {
	config := make(map[string]string)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return config
	}
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("cant open file:", err)
		return config
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			config[key] = value
		}
	}
	return config
}

func apply_global() bool {
	fmt.Println("[*] Applying Global Settings...")
	forward := "0"

	file, err := os.Open(global_conf)
	if err == nil {

		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "ip_forward=1") {
				forward = "1"
				break
			}
		}
	}
	run("sysctl -w net.ipv4.ip_forward=" + forward)
	return forward == "1"

}

func generate_subnet(index int) string {
	return fmt.Sprintf("10.0.%d.1", 50+index)
}

func main() {
	clean_slate()
	ip_forwarding_enabled := apply_global()
	intent_files, _ := filepath.Glob(intent_dir + "/*.conf")

}
