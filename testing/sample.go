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
	Ipv4 string
}

type network_config struct {
	Role            string
	Wireless_status string
	Ps              string
	Ssid            string
	Subnet          string
}

const intent_dir string = "/etc/net/interfaces"
const global_conf string = "/etc/net/global.conf"
const enforcer string = "/usr/lib/net/enforce.py"

func dump_global(filepa string, config global_config) {
	file, err := os.OpenFile(filepa, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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
	file, err := os.OpenFile(filepa, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
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
	final := []string{}
	slice := strings.Split(text, "\n")
	for _, line := range slice {
		final = append(final, strings.TrimSpace(strings.Split(line, ":")[1]))
	}
	return final
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
				fmt.Println("directory couldnt be created(global):", err)
				return
			}
			con := load_global(global_conf)
			if con.Ipv4 == "yes" {
				fmt.Println("ipv4 forwarding  is on")
				con.Ipv4 = "no"
			} else {
				fmt.Println("ipv4 forwarding  is off")
				con.Ipv4 = "yes"
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
			}
			var choice3 int
			fmt.Println("your option (0,1,2 ..)")
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
			var netwo_conf network_config
			if choice4 == 1 {
				netwo_conf.Role = "uplink"
				if strings.HasPrefix(curr_iface, "w") {
					netwo_conf.Wireless_status = "y"

					var choice5 string
					fmt.Println("would you like to scan for access points (y/n)")
					fmt.Scanln(&choice5)
					if choice5 == "y" {
						wifi_slice := scan_wifi(curr_iface)
						for index, ssid := range wifi_slice {
							fmt.Println(index+1, ssid)
						}
						var choice6 int
						fmt.Println("your option (1,2,3 ..)")
						fmt.Scanln(&choice6)
						var choice7 string
						fmt.Println("enter password")
						fmt.Scanln(&choice7)

						netwo_conf.Ssid = wifi_slice[choice6-1]
						netwo_conf.Ps = choice7
					} else {
						var choice6 string
						fmt.Println("enter ssid manually")
						fmt.Scanln(&choice6)
						var choice7 string
						fmt.Println("enter password")
						fmt.Scanln(&choice7)
						netwo_conf.Ps = choice7
						netwo_conf.Ssid = choice6
					}

				} else {
					netwo_conf.Wireless_status = "n"
				}
			} else if choice4 == 2 {
				netwo_conf.Role = "downlink"
				if strings.HasPrefix(curr_iface, "w") {
					var choice6 string
					fmt.Println("enter ssid for hotspot")
					fmt.Scanln(&choice6)
					var choice7 string
					fmt.Println("enter password")
					fmt.Scanln(&choice7)
					netwo_conf.Ps = choice7
					netwo_conf.Ssid = choice6

				} else {
					var choice6 string
					fmt.Println("enter subnet for downlink")
					fmt.Scanln(&choice6)
					netwo_conf.Subnet = choice6
				}

			} else if choice4 == 3 {
				netwo_conf.Role = "unmanaged"
			} else {

			}
			err := os.MkdirAll(intent_dir, 0755)
			if err != nil {
				fmt.Println("directory couldnt be created(intent):", err)
				return
			}
			network_file := intent_dir + "/" + curr_iface + ".conf"
			dump_network(network_file, netwo_conf)
		} else {
			break
		}
	}

}
