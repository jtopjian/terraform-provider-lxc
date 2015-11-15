package lxc

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vishvananda/netlink"
)

func TestLXCBridge(t *testing.T) {
	var bridge netlink.Link
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLXCBridgeDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccLXCBridge,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLXCBridgeExists(
						t, "lxc_bridge.accept_test", &bridge),
					resource.TestCheckResourceAttr(
						"lxc_bridge.accept_test", "name", "accept_test"),
				),
			},
			resource.TestStep{
				Config: testAccLXCBridgeWithIface,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLXCBridgeExists(
						t, "lxc_bridge.accept_test_iface", &bridge),
					resource.TestCheckResourceAttr(
						"lxc_bridge.accept_test_iface", "name", "accept_test_iface"),
					resource.TestCheckResourceAttr(
						"lxc_bridge.accept_test_iface", "hostInterface", "accept_test"),
				),
			},
		},
	})
}

func testAccCheckLXCBridgeExists(t *testing.T, n string, bridge *netlink.Link) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %v", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		bridgeIndex, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Internal error reading resource ID.")
		}
		br, err := netlink.LinkByIndex(bridgeIndex)
		if err != nil {
			return fmt.Errorf("Error searching for bridge.")
		} else {
			*bridge = br
			return nil
		}

		return fmt.Errorf("Unable to find bridge.")
	}
}

func testAccCheckLXCBridgeDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "lxc_bridge" {
			continue
		}

		bridgeIndex, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Internal error reading resource ID.")
		}
		_, err = netlink.LinkByIndex(bridgeIndex)
		if err == nil {
			return fmt.Errorf("Bridge still exists.")
		}

	}

	return nil
}

var testAccLXCBridge = `
	resource "lxc_bridge" "accept_test" {
		name = "accept_test"
	}`

var testAccLXCBridgeWithIface = `
	resource "lxc_bridge" "accept_test" {
		name = "accept_test_ip"
		hostInterface = "accept_test"
	}`
