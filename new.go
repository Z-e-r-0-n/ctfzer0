// check whether the interface is there before implementing the intent files
// denforce.py doesnt do downlink via wired implement that too

package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

type global_config struct {
	Ipv4 string
}

type network_config struct {
	Role            string
	Wireless_status string
	Ps              string
	Ssid            string
	Subnet          string
	Name string
}

func get_interfaces() []string {
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("get interfaces failed", err)
		return []string{}
	}
	interfac := []string{}
	for _, iface := range interfaces {
		if strings.Contains(iface.Name, "vir") || iface.Name == "lo" {
			continue
		}
		interfac = append(interfac, iface.Name)
	}
	return interfac

}

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

func clean_slate(interfaces []string) {
	fmt.Println("[*] Cleaning old configurations...")

	files, _ := filepath.Glob(systemd_net_dir + "/*.network")

	for _, f := range files {
		baseName := filepath.Base(f)
		if !slices.Contains(interfaces, strings.Split(baseName, ".")[0]) {
			err := os.Remove(f)
			if err != nil {
				fmt.Println("failed to remove :", f, err)
			}
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

func load_network(filepa string) network_config {
	file, err := os.OpenFile(filepa, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println("file error", err)
		return network_config{}
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	var unpickled network_config
	err = decoder.Decode(&unpickled)
	return unpickled
}

func apply_global() {
	file, err := os.OpenFile(global_conf, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println("file error", err)
		return
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	var unpickled global_config
	err = decoder.Decode(&unpickled)
	var forward string
	if unpickled.Ipv4 == "yes" {
		forward = "1"
	} else {
		forward = "0"

	}
	run("sysctl -w net.ipv4.ip_forward=" + forward)

}

func generate_subnet(index int) string {
	return fmt.Sprintf("10.0.%d.1", 50+index)
}

func main() {
	interfaces := get_interfaces()
	clean_slate(interfaces)
	apply_global()
	intent_files, _ := filepath.Glob(intent_dir + "/*.conf")

	var final_intents []string
	for _, file := range intent_files {
		baseName := filepath.Base(file)
		filet := strings.Split(baseName, ".")[0]
		if slices.Contains(interfaces, filet) {
			final_intents = append(final_intents, file)
		}
	}

	for _ , final := range(final_intents){
		current_intent := load_network(final)
		if current_intent.Role == "uplink"{
			metric := 200 
            if strings.Contains(current_intent.Name , "et"){ 
				metric = 100
			}
            if strings.Contains(current_intent.Name , "en"){ 
				metric = 100
			}
			if strings.Contains(current_intent.Name , "wl"){ 
				metric = 100
			}
			
		}



	}

}
