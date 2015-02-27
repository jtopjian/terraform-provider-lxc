package lxc

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
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

func setConfigItems(c *lxc.Container, configPath string, options map[string]interface{}) error {

	// first append a line to the main config file that points to another.
	// check and see if that line has been appended already.
	// if not, append it.
	configFile := configPath + "/" + c.Name() + "/config"
	customConfigFile := configPath + "/" + c.Name() + "/config_tf"
	customConfigLine := fmt.Sprintf("lxc.include = %s", customConfigFile)

	log.Printf("[DEBUG] %v", configFile)
	configFileContents, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}

	configFound := false
	lines := strings.Split(string(configFileContents), "\n")
	for _, line := range lines {
		if line == customConfigLine {
			configFound = true
		}
	}

	// if the lxc.include line was not found, add it.
	if configFound == false {
		lines = append(lines, customConfigLine)
		lines = append(lines, "\n")
		log.Printf("[DEBUG] %v", lines)
		if err := ioutil.WriteFile(configFile, []byte(strings.Join(lines, "\n")), 0640); err != nil {
			return err
		}
	}

	// now rewrite all custom config options
	lines = lines[:0]
	for k, v := range options {
		lines = append(lines, fmt.Sprintf("%s = %s", k, v.(string)))
	}
	if err := ioutil.WriteFile(customConfigFile, []byte(strings.Join(lines, "\n")), 0640); err != nil {
		return err
	}

	return nil
}

func checkBackend(backend string) (lxc.BackendStore, error) {
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
