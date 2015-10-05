package lxc

import (
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vishvananda/netlink"
)

func resourceLXCBridge() *schema.Resource {
	return &schema.Resource{
		Create: resourceLXCBridgeCreate,
		Read:   resourceLXCBridgeRead,
		Update: nil,
		Delete: resourceLXCBridgeDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"hostInterface": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"mac": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceLXCBridgeCreate(d *schema.ResourceData, meta interface{}) error {
	br := d.Get("name").(string)
	var bridge netlink.Link

	if bridge, err := netlink.LinkByName(br); err != nil {
		bridge = &netlink.Bridge{netlink.LinkAttrs{
			Name: d.Get("name").(string),
		}}
		if err := netlink.LinkAdd(bridge); err != nil {
			return fmt.Errorf("Error creating bridge %s: %v", br, err)
		}

		if ifaceName, ok := d.GetOk("hostInterface"); ok {
			iface, err := netlink.LinkByName(ifaceName.(string))
			if err != nil {
				return fmt.Errorf("Error adding host interface %s to bridge %s : unknow host interface %v", ifaceName, br ,err)
			}

			if err := netlink.LinkSetMasterByIndex(iface, bridge.Attrs().Index); err != nil {
				return fmt.Errorf("Error adding host interface %s to bridge %s : %v", ifaceName, br, err)
			}
		}
		log.Printf("[INFO] Created new bridge %s: %v", br, bridge)
	} else {
		log.Printf("[INFO] Found existing bridge %s: %v", br, bridge)
	}

	log.Printf("[INFO] Bringing bridge up.")
	if err := netlink.LinkSetUp(bridge); err != nil {
		return fmt.Errorf("Error bringing bridge up: %v", err)
	}

	d.SetId(strconv.Itoa(bridge.Attrs().Index))

	return resourceLXCBridgeRead(d, meta)
}

func resourceLXCBridgeRead(d *schema.ResourceData, meta interface{}) error {
	bridgeIndex, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Internal error reading resource ID: %v", err)
	}

	bridge, err := netlink.LinkByIndex(bridgeIndex)
	if err != nil {
		return fmt.Errorf("Unable to find bridge %v: %v", bridgeIndex, err)
	}

	d.Set("mac", bridge.Attrs().HardwareAddr.String())

	log.Printf("[INFO] Bridge info: %v", bridge)

	return nil
}

func resourceLXCBridgeDelete(d *schema.ResourceData, meta interface{}) error {
	bridgeIndex, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Internal error reading resource ID: %v", err)
	}

	bridge, err := netlink.LinkByIndex(bridgeIndex)
	if err != nil {
		return fmt.Errorf("Unable to find bridge %v: %v", bridgeIndex, err)
	}

	links, err := netlink.LinkList()
	if err != nil {
		return fmt.Errorf("Error listing interfaces: %v", err)
	}

	bridgeEmpty := true
	for _, link := range links {
		if link.Attrs().MasterIndex == bridge.Attrs().Index {
			bridgeEmpty = false
			log.Printf("[INFO] Link %s is still attached to bridge %s", link.Attrs().Name, bridge.Attrs().Name)
		}
	}

	if bridgeEmpty == false {
		return fmt.Errorf("Unable to delete bridge %s. Interfaces are still attached to it.", bridge.Attrs().Name)
	} else {
		if err := netlink.LinkDel(bridge); err != nil {
			return fmt.Errorf("Error deleting bridge: %s", err)
		}
	}

	return nil
}
