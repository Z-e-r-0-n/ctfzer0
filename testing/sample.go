package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type global_config struct {
	ipv4 string
}

type network_config struct {
}

const intent_dir string = "/etc/net/interfaces"
const global_conf string = "/etc/net/global.conf"
const enforcer string = "/usr/lib/net/enforce.py"

func dump_global(filepa string, config global_config) {
	file, err := os.OpenFile(filepa, os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("file error", err)
		return
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(config)
	if err != nil {
		fmt.Println("Error encoding binary GLOB:", err)
	}
}
func load_global(filepa string) global_config {
	file, err := os.OpenFile(filepa, os.O_RDONLY, 0644)
	if err != nil {
		fmt.Println("file error", err)
		return global_config{}
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	var unpickled global_config
	err = decoder.Decode(&unpickled)
	return unpickled
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

func dump_network(filepa string, config network_config) {
	file, err := os.OpenFile(filepa, os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("file error", err)
		return
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(config)
	if err != nil {
		fmt.Println("Error encoding binary NEt:", err)
	}
}
func get_interfaces() []string {
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("dns resolve failed:", err)
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

func scan_wifi(interfa string) []string {
	cmdstring1 := fmt.Sprintf("ip link set %s up ", interfa)
	cmd1 := exec.Command("sh", "-c", cmdstring1)
	cmdstring2 := fmt.Sprintf("iwlist %s scan | grep ESSID", interfa)
	cmd2 := exec.Command("sh", "-c", cmdstring2)
	_, err := cmd1.CombinedOutput()
	if err != nil {
		fmt.Println("link couldnt be set up:", err)
	}

	output, err := cmd2.CombinedOutput()
	if err != nil {
		fmt.Println("wifi scan failed:", err)
	}
	text := string(output)
	slice := strings.Split(text, ":")
	return slice
}

func main() {
	if os.Geteuid() != 0 {
		fmt.Println("Run with higher privillge..")
		return
	}
	for {
		fmt.Println("1. Global settings \n", "2. interface settings \n", "3. exit ")
		var choice int
		fmt.Print("Enter choice: ")
		fmt.Scanln(&choice)
		if choice == 1 {
			dir_path := filepath.Dir(global_conf)
			err := os.MkdirAll(dir_path, 0755)
			if err != nil {
				fmt.Println("directory couldnt be created:", err)
				return
			}
			con := load_global(global_conf)
			if con.ipv4 == "yes" {
				fmt.Println("ipv4 forwarding  is on")
				con.ipv4 = "no"
			} else {
				fmt.Println("ipv4 forwarding  is off")
				con.ipv4 = "yes"
			}
			var choice2 string
			fmt.Print("do you wish to change y/n ?")
			fmt.Scanln(&choice2)
			if choice2 == "y" {
				dump_global(global_conf, con)
			}
		} else if choice == 2 {
			interfaces := get_interfaces()
			for index, iface := range interfaces {
				fmt.Println(index, ".", iface)
				var choice3 int
				fmt.Println("your option (1,2,3 ..)")
				fmt.Scanln(&choice3)
				curr_iface := string(interfaces[choice3])
				fmt.Println("configuring ", curr_iface)
				fmt.Println("1. Uplink")
				fmt.Println("2. Downlink")
				fmt.Println("3. Unmanaged")
				fmt.Println("4. Clear Configuration")
				var choice4 int
				fmt.Println("your option (1,2,3 ..)")
				fmt.Scanln(&choice4)
				if choice4 == 1 {
					if curr_iface == "w" {
						var choice5 string
						fmt.Println("would you like to scan for access points (y/n)")
						fmt.Scanln(&choice5)
						if choice5 == "y" {
							wifi_slice := scan_wifi(curr_iface)
							fmt.Println(wifi_slice)
						}

					} else {

					}
				}

			}
		} else {
			break
		}

	}
}
