package lxc

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"lxc_path": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/var/lib/lxc",
			},
			"lxc_log_path": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "/var/log/lxc",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"lxc_container": resourceLXCContainer(),
			"lxc_clone":     resourceLXCClone(),
		},

		ConfigureFunc: configureProvider,
	}
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {

	config := Config{
		LXCPath:    d.Get("lxc_path").(string),
		LXCLogPath: d.Get("lxc_log_path").(string),
	}

	return &config, nil
}
