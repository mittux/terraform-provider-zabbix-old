package provider

import (
	logger "log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/tpretz/go-zabbix-api"
)

// Provider definition
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Zabbix API username",
				ValidateFunc: validation.StringIsNotWhiteSpace,
				DefaultFunc:  schema.MultiEnvDefaultFunc([]string{"ZABBIX_USER", "ZABBIX_USERNAME"}, nil),
			},
			"password": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Zabbix API password",
				ValidateFunc: validation.StringIsNotWhiteSpace,
				DefaultFunc:  schema.MultiEnvDefaultFunc([]string{"ZABBIX_PASS", "ZABBIX_PASSWORD"}, nil),
			},
			"url": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Zabbix API url",
				DefaultFunc:  schema.MultiEnvDefaultFunc([]string{"ZABBIX_URL", "ZABBIX_SERVER_URL"}, nil),
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			},
			"tls_insecure": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Disable TLS certificate checking (for testing use only)",
				Optional:    true,
				Default:     false,
			},
			"serialize": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Serialize API requests, if required due to API race conditions",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"zabbix_host":        dataHost(),
			"zabbix_proxy":       dataProxy(),
			"zabbix_hostgroup":   dataHostgroup(),
			"zabbix_template":    dataTemplate(),
			"zabbix_application": dataApplication(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"zabbix_item_trapper":   resourceItemTrapper(),
			"zabbix_item_http":      resourceItemHttp(),
			"zabbix_item_simple":    resourceItemSimple(),
			"zabbix_item_internal":  resourceItemInternal(),
			"zabbix_item_snmp":      resourceItemSnmp(),
			"zabbix_item_agent":     resourceItemAgent(),
			"zabbix_item_aggregate": resourceItemAggregate(),
			"zabbix_item_dependent": resourceItemDependent(),
			"zabbix_application":    resourceApplication(),
			"zabbix_trigger":        resourceTrigger(),
			"zabbix_template":       resourceTemplate(),
			"zabbix_hostgroup":      resourceHostgroup(),
			"zabbix_host":           resourceHost(),
		},
		ConfigureFunc: providerConfigure,
	}
}

// providerConfigure configure this provider
func providerConfigure(d *schema.ResourceData) (meta interface{}, err error) {
	log.Trace("Started zabbix provider init")
	l := logger.New(stderr, "[DEBUG] ", logger.LstdFlags)

	api := zabbix.NewAPI(zabbix.Config{
		Url:         d.Get("url").(string),
		TlsNoVerify: d.Get("tls_insecure").(bool),
		Log:         l,
		Serialize:   d.Get("serialize").(bool),
	})

	_, err = api.Login(d.Get("username").(string), d.Get("password").(string))
	meta = api
	log.Trace("Started zabbix provider got error: %+v", err)

	return
}
