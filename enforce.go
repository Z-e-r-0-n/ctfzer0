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
	Name            string
}

const intent_dir string = "/etc/net/interfaces"
const global_conf string = "/etc/net/global.conf"

const systemd_net_dir string = "/etc/systemd/network"
const wpa_supplicant_dir string = "/etc/wpa_supplicant"
const hostapd_conf_dir = "/etc/hostapd"
const dnsmasq_conf_dir = "/etc/dnsmasq.d"

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

func write_file(path string, content string) error {
	err := os.WriteFile(path, []byte(content), 0600)
	if err != nil {
		return err
	}
	return nil
}

func run(command string) {
	c := exec.Command("sh", "-c", command)

	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err := c.Run()
	if err != nil {
		fmt.Println("Command failed :", command, "->", err)
	}
}

func clean_slate(interfaces []string) {
	fmt.Println("[*] Cleaning old configurations...")

	files, _ := filepath.Glob(systemd_net_dir + "/*.network")
	for _, f := range files {
		baseName := filepath.Base(f)
		shouldKeep := false
		for _, iface := range interfaces {
			if strings.Contains(baseName, iface) {
				shouldKeep = true
				break
			}
		}
		if !shouldKeep {
			os.Remove(f)
		}
	}

	run("systemctl stop hostapd")
	run("pkill hostapd")
	run("systemctl stop dnsmasq")

	hostapdFiles, _ := filepath.Glob(hostapd_conf_dir + "/*.conf")
	for _, f := range hostapdFiles {
		os.Remove(f)
	}

	dnsmasqFiles, _ := filepath.Glob(dnsmasq_conf_dir + "/*.conf")
	for _, f := range dnsmasqFiles {
		os.Remove(f)
	}

	run("pkill wpa_supplicant")
}

func load_network(filepa string) network_config {
	file, err := os.OpenFile(filepa, os.O_RDONLY, 0644)
	if err != nil {
		return network_config{}
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	var unpickled network_config
	decoder.Decode(&unpickled)
	return unpickled
}

func apply_global() bool {
	fmt.Println("applying global settings.")
	file, err := os.OpenFile(global_conf, os.O_RDONLY, 0644)
	if err != nil {
		run("sysctl -w net.ipv4.ip_forward=0")
		return false
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	var unpickled global_config
	decoder.Decode(&unpickled)

	if unpickled.Ipv4 == "yes" {
		run("sysctl -w net.ipv4.ip_forward=1")
		return true
	}
	run("sysctl -w net.ipv4.ip_forward=0")
	return false
}

func generate_subnet(index int) string {
	return fmt.Sprintf("10.0.%d.1", 50+index)
}

func main() {
	if os.Geteuid() != 0 {
		fmt.Println("Run as root.")
		return
	}
	interfaces := get_interfaces()
	clean_slate(interfaces)
	ip_forwarding_enabled := apply_global()
	intent_files, _ := filepath.Glob(intent_dir + "/*.conf")
	var final_intents []string
	for _, file := range intent_files {
		baseName := filepath.Base(file)
		filet := strings.Split(baseName, ".")[0]
		if slices.Contains(interfaces, filet) {
			final_intents = append(final_intents, file)
		}
	}
	var uplink_interfaces []string
	downlink_index := 0
	for _, final := range final_intents {
		current_intent := load_network(final)
		iface := strings.Split(filepath.Base(final), ".")[0]
		switch current_intent.Role {
		case "uplink":
			uplink_interfaces = append(uplink_interfaces, iface)
			metric := 200
			if strings.Contains(iface, "et") || strings.Contains(iface, "en") {
				metric = 100
			}
			if strings.Contains(iface, "wl") {
				metric = 300
			}
			fmt.Printf("configuring %s as UPLINK (Metric: %d)\n", iface, metric)
			net_content := fmt.Sprintf("[Match]\nName=%s\n\n[Network]\nDHCP=yes\nIPMasquerade=yes\n\n[DHCPv4]\nRouteMetric=%d\n", iface, metric)
			write_file(fmt.Sprintf("%s/%s.network", systemd_net_dir, iface), net_content)
			if current_intent.Wireless_status == "y" && current_intent.Ssid != "" {
				wpa_content := fmt.Sprintf("ctrl_interface=/run/wpa_supplicant\nupdate_config=1\ncountry=IN\n\nnetwork={\n    ssid=\"%s\"\n    psk=\"%s\"\n}\n", current_intent.Ssid, current_intent.Ps)
				write_file(fmt.Sprintf("%s/wpa_supplicant-%s.conf", wpa_supplicant_dir, iface), wpa_content)

				run(fmt.Sprintf("systemctl enable wpa_supplicant@%s", iface))
				run(fmt.Sprintf("systemctl restart wpa_supplicant@%s", iface))
			}
		case "downlink":
			fmt.Printf("configuring %s as DOWNLINK\n", iface)
			gw_ip := current_intent.Subnet
			if gw_ip == "" {

				gw_ip = generate_subnet(downlink_index)
				downlink_index++
			}
			ipParts := strings.Split(gw_ip, ".")
			dhcp_range := ""
			if len(ipParts) == 4 {
				baseIP := fmt.Sprintf("%s.%s.%s", ipParts[0], ipParts[1], ipParts[2])
				dhcp_range = fmt.Sprintf("%s.10,%s.100,12h", baseIP, baseIP)
			}
			net_content := fmt.Sprintf("[Match]\nName=%s\n\n[Network]\nAddress=%s/24\nIPMasquerade=yes\nDHCPServer=no\n", iface, gw_ip)
			write_file(fmt.Sprintf("%s/%s.network", systemd_net_dir, iface), net_content)
			if current_intent.Wireless_status == "y" && current_intent.Ssid != "" {
				hostapd_conf := fmt.Sprintf("%s/%s.conf", hostapd_conf_dir, iface)
				hostapd_content := fmt.Sprintf("interface=%s\ndriver=nl80211\nssid=%s\nhw_mode=g\nchannel=7\nwmm_enabled=0\nmacaddr_acl=0\nauth_algs=1\nignore_broadcast_ssid=0\nwpa=2\nwpa_passphrase=%s\nwpa_key_mgmt=WPA-PSK\nwpa_pairwise=TKIP\nrsn_pairwise=CCMP\n", iface, current_intent.Ssid, current_intent.Ps)
				write_file(hostapd_conf, hostapd_content)

				run("systemctl unmask hostapd")
				run(fmt.Sprintf("hostapd -B %s", hostapd_conf))
			}
			dnsmasq_content := fmt.Sprintf("interface=%s\nbind-interfaces\ndhcp-range=%s\ndhcp-option=3,%s\ndhcp-option=6,%s\nserver=127.0.0.53\n", iface, dhcp_range, gw_ip, gw_ip)
			write_file(fmt.Sprintf("%s/%s.conf", dnsmasq_conf_dir, iface), dnsmasq_content)
			downlink_index++

		case "unmanaged":
			fmt.Printf("configuring %s as UNMANAGED\n", iface)

			run(fmt.Sprintf("ip link set %s up", iface))
		}
		fmt.Printf("enforcement complete")
		os.Remove(final)
	}
	fmt.Println("reloading systemd-networkd...")
	run("systemctl daemon-reload")
	run("systemctl restart systemd-networkd")
	if downlink_index > 0 {
		fmt.Println("starting DNS/DHCP services...")
		run("systemctl restart dnsmasq")
	}
	if ip_forwarding_enabled && len(uplink_interfaces) > 0 {
		fmt.Println("enabling NAT rules for uplinks...")
		for _, up_iface := range uplink_interfaces {
			run(fmt.Sprintf("iptables -t nat -A POSTROUTING -o %s -j MASQUERADE", up_iface))
		}
	}
	fmt.Println("network state enforced successfully.")
}
