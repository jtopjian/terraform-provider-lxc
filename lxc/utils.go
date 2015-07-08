package lxc

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"gopkg.in/lxc/go-lxc.v2"
)

func lxcContainerStateRefreshFunc(name, lxcpath string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		c, err := lxc.NewContainer(name, lxcpath)
		if err != nil {
			return c, "", err
		}
		state := c.State()
		return c, fmt.Sprintf("%s", state), nil
	}
}

func lxcWaitForState(c *lxc.Container, LXCPath string, pendingStates []string, targetState string) error {
	stateConf := &resource.StateChangeConf{
		Pending:    pendingStates,
		Target:     targetState,
		Refresh:    lxcContainerStateRefreshFunc(c.Name(), LXCPath),
		Timeout:    10 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 1 * time.Second,
	}

	_, err := stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("Error waiting for container (%s) to change to state (%s): %s", c.Name(), targetState, err)
	}

	return nil
}

func lxcOptions(c *lxc.Container, d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	var options []string
	optionsFound := false
	includeFound := false
	configFile := config.LXCPath + "/" + c.Name() + "/config"
	customConfigFile := config.LXCPath + "/" + c.Name() + "/config_tf"
	includeLine := fmt.Sprintf("lxc.include = %s", customConfigFile)

	networkInterfaces := d.Get("network_interface").([]interface{})
	for _, n := range networkInterfaces {
		nic := n.(map[string]interface{})
		options = append(options, fmt.Sprintf("lxc.network.type = %s", nic["type"]))
		for k, v := range nic["options"].(map[string]interface{}) {
			options = append(options, fmt.Sprintf("lxc.network.%s = %s", k, v.(string)))
		}
	}

	containerOptions := d.Get("options").(map[string]interface{})
	if containerOptions != nil {
		optionsFound = true
		for k, v := range containerOptions {
			options = append(options, fmt.Sprintf("%s = %s", k, v.(string)))
		}
	}

	configFileContents, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	if optionsFound == true {
		lines := strings.Split(string(configFileContents), "\n")
		for _, line := range lines {
			if line == includeLine {
				includeFound = true
			}
		}

		// if the lxc.include line was not found, add it.
		if includeFound == false {
			lines = append(lines, includeLine, "\n")
			if err := ioutil.WriteFile(configFile, []byte(strings.Join(lines, "\n")), 0640); err != nil {
				return err
			}
		}

		// now rewrite all custom config options
		log.Printf("[DEBUG] %v", options)
		if err := ioutil.WriteFile(customConfigFile, []byte(strings.Join(options, "\n")), 0640); err != nil {
			return err
		}
	}

	return nil
}

func lxcCheckBackend(backend string) (lxc.BackendStore, error) {
	switch backend {
	case "btrfs":
		return lxc.Btrfs, nil
	case "directory":
		return lxc.Directory, nil
	case "lvm":
		return lxc.LVM, nil
	case "zfs":
		return lxc.ZFS, nil
	case "aufs":
		return lxc.Aufs, nil
	case "overlayfs":
		return lxc.Overlayfs, nil
	case "loopback":
		return lxc.Loopback, nil
	case "best":
		return lxc.Best, nil
	default:
		return 0, fmt.Errorf("Invalid backend. Possible values are: btrfs, directory, lvm, zfs, aufs, overlayfs, loopback, or best.")
	}
}

func lxcIPAddressConfiguration(c *lxc.Container, d *schema.ResourceData) error {
	// Loop through all interfaces and see if one is marked as management
	managementNIC := "eth0"
	i := 0
	networkInterfaces := d.Get("network_interface").([]interface{})
	for _, n := range networkInterfaces {
		nic := n.(map[string]interface{})
		if nic["management"] == true {
			managementNIC = fmt.Sprintf("%s%s", "eth", strconv.Itoa(i))
		}
		i++
	}

	// Get the IP addresses of the management NIC
	// For now, we'll just use the first returned IP.
	d.Set("address_v4", "")
	ipv4s, err := c.IPv4Address(managementNIC)
	if err == nil {
		if len(ipv4s) > 0 {
			d.Set("address_v4", ipv4s[0])
			d.SetConnInfo(map[string]string{
				"type": "ssh",
				"host": ipv4s[0],
			})
		}
	}

	d.Set("address_v6", "")
	ipv6s, err := c.IPv6Address(managementNIC)
	if err == nil {
		if len(ipv6s) > 0 {
			d.Set("address_v6", ipv6s[0])
		}
	}

	return nil
}
