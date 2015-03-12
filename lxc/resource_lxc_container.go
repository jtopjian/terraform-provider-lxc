package lxc

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"gopkg.in/lxc/go-lxc.v2"
)

func resourceLXCContainer() *schema.Resource {
	return &schema.Resource{
		Create: resourceLXCContainerCreate,
		Read:   resourceLXCContainerRead,
		Update: nil,
		Delete: resourceLXCContainerDelete,

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
			"template_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "download",
				ForceNew: true,
			},
			"template_distro": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "ubuntu",
				ForceNew: true,
			},
			"template_release": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "trusty",
				ForceNew: true,
			},
			"template_arch": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "amd64",
				ForceNew: true,
			},
			"template_variant": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "default",
				ForceNew: true,
			},
			"template_server": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "images.linuxcontainers.org",
				ForceNew: true,
			},
			"template_key_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"template_key_server": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"template_flush_cache": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"template_force_cache": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"template_disable_gpg_validation": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"template_extra_args": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set: func(v interface{}) int {
					return hashcode.String(v.(string))
				},
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

func resourceLXCContainerCreate(d *schema.ResourceData, meta interface{}) error {
	var c *lxc.Container
	config := meta.(*Config)

	backendType, err := lxcCheckBackend(d.Get("backend").(string))
	if err != nil {
		return err
	}

	name := d.Get("name").(string)
	c, err = lxc.NewContainer(name, config.LXCPath)
	if err != nil {
		return err
	}

	d.SetId(c.Name())

	log.Printf("[INFO] Creating container %s\n", c.Name())

	var ea []string
	for _, v := range d.Get("template_extra_args").(*schema.Set).List() {
		ea = append(ea, v.(string))
	}

	var options lxc.TemplateOptions
	templateName := d.Get("template_name").(string)
	if templateName == "download" {
		options = lxc.TemplateOptions{
			Backend:              backendType,
			Template:             d.Get("template_name").(string),
			Distro:               d.Get("template_distro").(string),
			Release:              d.Get("template_release").(string),
			Arch:                 d.Get("template_arch").(string),
			Variant:              d.Get("template_variant").(string),
			Server:               d.Get("template_server").(string),
			KeyID:                d.Get("template_key_id").(string),
			KeyServer:            d.Get("template_key_server").(string),
			FlushCache:           d.Get("template_flush_cache").(bool),
			ForceCache:           d.Get("template_force_cache").(bool),
			DisableGPGValidation: d.Get("template_disable_gpg_validation").(bool),
			ExtraArgs:            ea,
		}
	} else {
		options = lxc.TemplateOptions{
			Backend:    backendType,
			Template:   d.Get("template_name").(string),
			Release:    d.Get("template_release").(string),
			Arch:       d.Get("template_arch").(string),
			FlushCache: d.Get("template_flush_cache").(bool),
			ExtraArgs:  ea,
		}
	}

	if err := c.Create(options); err != nil {
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

	return resourceLXCContainerRead(d, meta)
}

func resourceLXCContainerRead(d *schema.ResourceData, meta interface{}) error {
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

	d.SetConnInfo(map[string]string{
		"type": "ssh",
		"host": ipv4,
	})

	return nil
}

func resourceLXCContainerDelete(d *schema.ResourceData, meta interface{}) error {
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
