package lxc

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"

	"gopkg.in/lxc/go-lxc.v2"
)

func TestLXCClone(t *testing.T) {
	var container lxc.Container
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLXCCloneDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccLXCClone,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLXCCloneExists(
						t, "lxc_clone.accept_clone", &container),
					resource.TestCheckResourceAttr(
						"lxc_clone.accept_clone", "name", "accept_clone"),
				),
			},
		},
	})
}

func testAccCheckLXCCloneExists(t *testing.T, n string, container *lxc.Container) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %v", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config, err := testProviderConfig()
		if err != nil {
			return err
		}

		c := lxc.ActiveContainers(config.LXCPath)
		for i := range c {
			if c[i].Name() == rs.Primary.ID {
				*container = c[i]
				return nil
			}
		}

		return fmt.Errorf("Unable to find running container.")
	}
}

func testAccCheckLXCCloneDestroy(s *terraform.State) error {
	config, err := testProviderConfig()
	if err != nil {
		return err
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "lxc_container" {
			continue
		}

		c := lxc.ActiveContainers(config.LXCPath)
		for i := range c {
			if c[i].Name() == rs.Primary.ID {
				return fmt.Errorf("Container still exists.")
			}
		}
	}

	return nil
}

var testAccLXCClone = `
	resource "lxc_container" "accept_test" {
		name = "accept_test"
	}

	resource "lxc_clone" "accept_clone" {
		name = "accept_clone"
		source = "${lxc_container.accept_test.name}"
	}`
