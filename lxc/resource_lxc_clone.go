package lxc

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"gopkg.in/lxc/go-lxc.v2"
)

func resourceLXCClone() *schema.Resource {
	return &schema.Resource{
		Create: resourceLXCCloneCreate,
		Read:   resourceLXCCloneRead,
		Update: nil,
		Delete: resourceLXCCloneDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"backend": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "directory",
				ForceNew: true,
			},
			"source": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"keep_mac": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"snapshot": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"options": &schema.Schema{
				Type:     schema.TypeMap,
				Optional: true,
				Default:  nil,
			},
			"network_interface": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default:  "veth",
						},
						"options": &schema.Schema{
							Type:     schema.TypeMap,
							Optional: true,
							Default:  nil,
						},
					},
				},
			},

			// exported
			"address_v4": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"address_v6": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceLXCCloneCreate(d *schema.ResourceData, meta interface{}) error {
	var c *lxc.Container
	config := meta.(*Config)

	backendType, err := lxcCheckBackend(d.Get("backend").(string))
	if err != nil {
		return err
	}

	name := d.Get("name").(string)
	source := d.Get("source").(string)

	c, err = lxc.NewContainer(name, config.LXCPath)
	if err != nil {
		return err
	}

	d.SetId(c.Name())

	cl, err := lxc.NewContainer(source, config.LXCPath)
	if err != nil {
		return err
	}

	// the source container must be stopped
	log.Printf("[INFO] Stopping %s", source)
	if cl.State() == lxc.RUNNING {
		if err := cl.Stop(); err != nil {
			return err
		}
		if err := lxcWaitForState(c, config.LXCPath, []string{"RUNNING", "STOPPING"}, "STOPPED"); err != nil {
			return err
		}

	}

	log.Printf("[INFO] Cloning %s as %s", source, name)
	err = cl.Clone(name, lxc.CloneOptions{
		Backend:    backendType,
		ConfigPath: config.LXCPath,
		KeepMAC:    d.Get("keep_mac").(bool),
		Snapshot:   d.Get("snapshot").(bool),
	})
	if err != nil {
		return err
	}

	if err := lxcOptions(c, d, config); err != nil {
		return err
	}

	// causes lxc to re-read the config file
	c, err = lxc.NewContainer(name, config.LXCPath)
	if err != nil {
		return err
	}

	log.Printf("[INFO] Starting container %s\n", c.Name())
	if err := c.Start(); err != nil {
		return fmt.Errorf("Unable to start container: %s", err)
	}

	if err := lxcWaitForState(c, config.LXCPath, []string{"STOPPED", "STARTING"}, "RUNNING"); err != nil {
		return err
	}

	log.Printf("[INFO] Waiting container to startup networking...\n")
	c.WaitIPAddresses(5 * time.Second)

	return resourceLXCCloneRead(d, meta)
}

func resourceLXCCloneRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	c, err := lxc.NewContainer(d.Id(), config.LXCPath)
	if err != nil {
		return err
	}

	ipv4 := ""
	ipv6 := ""
	if ipv4s, err := c.IPv4Addresses(); err == nil {
		for _, v := range ipv4s {
			if ipv4 == "" {
				ipv4 = v
			}
		}
	}
	if ipv6s, err := c.IPv6Addresses(); err == nil {
		for _, v := range ipv6s {
			if ipv6 == "" {
				ipv6 = v
			}
		}
	}

	d.Set("address_v4", ipv4)
	d.Set("address_v6", ipv6)

	return nil
}

func resourceLXCCloneDelete(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)
	c, err := lxc.NewContainer(d.Id(), config.LXCPath)
	if err != nil {
		return err
	}

	if c.State() == lxc.RUNNING {
		if err := c.Stop(); err != nil {
			return err
		}

		if err := lxcWaitForState(c, config.LXCPath, []string{"RUNNING", "STOPPING"}, "STOPPED"); err != nil {
			return err
		}
	}

	if err := c.Destroy(); err != nil {
		return err
	}

	return nil
}
