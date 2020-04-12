package provider

import (
	"errors"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	zabbix "github.com/tpretz/go-zabbix-api"
)

// applicationSchemaBase base application schema
var applicationSchemaBase = map[string]*schema.Schema{
	"name": &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Optional:    false,
		Description: "Name of the application",
	},
	"hostid": &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		Description:  "Host ID",
		ValidateFunc: validation.StringMatch(regexp.MustCompile("^[0-9]+$"), "must be numeric"),
	},
}

// resourceApplication terraform application resource entrypoint
func resourceApplication() *schema.Resource {
	return &schema.Resource{
		Create: resourceApplicationCreate,
		Read:   resourceApplicationRead,
		Update: resourceApplicationUpdate,
		Delete: resourceApplicationDelete,
		Schema: applicationResourceSchema(applicationSchemaBase),
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
	}
}

// dataApplication terraform application resource entrypoint
func dataApplication() *schema.Resource {
	return &schema.Resource{
		Read:   dataApplicationRead,
		Schema: applicationDataSchema(applicationSchemaBase),
	}
}

// applicationResourceSchema adjust a base schema for resource usage
func applicationResourceSchema(m map[string]*schema.Schema) (o map[string]*schema.Schema) {
	o = map[string]*schema.Schema{}
	for k, v := range m {
		schema := *v

		// required
		switch k {
		case "name", "hostid":
			schema.Required = true
		}

		o[k] = &schema
	}

	return o
}

// applicationDataSchema adjust a base schema for data usage
func applicationDataSchema(m map[string]*schema.Schema) (o map[string]*schema.Schema) {
	o = map[string]*schema.Schema{}
	for k, v := range m {
		schema := *v

		// computed
		// switch k {
		// case "applicationid", "flags", "templateids":
		// 	schema.Optional = true
		// }

		o[k] = &schema
	}

	// lookup vars
	// o["hostid"] = &schema.Schema{
	// 	Type:     schema.TypeString,
	// 	Optional: true,
	// }

	return o
}

// buildApplicationObject create application struct
func buildApplicationObject(d *schema.ResourceData) (*zabbix.Application, error) {
	item := zabbix.Application{
		HostID: d.Get("hostid").(string),
		Name:   d.Get("name").(string),
	}

	log.Trace("build application object: %#v", item)

	return &item, nil
}

// resourceApplicationCreate terraform create handler
func resourceApplicationCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*zabbix.API)

	item, err := buildApplicationObject(d)
	if err != nil {
		return err
	}

	items := []zabbix.Application{*item}

	err = api.ApplicationsCreate(items)
	if err != nil {
		return err
	}

	log.Trace("created application: %+v", items[0])

	d.SetId(items[0].ApplicationID)

	return resourceApplicationRead(d, m)
}

// dataApplicationRead read handler for data resource
func dataApplicationRead(d *schema.ResourceData, m interface{}) error {
	params := zabbix.Params{
		"filter": map[string]interface{}{},
	}

	lookups := []string{"applicationid", "hostid", "name"}
	for _, k := range lookups {
		if v, ok := d.GetOk(k); ok {
			params["filter"].(map[string]interface{})[k] = v
		}
	}

	log.Debug("performing data lookup with params: %#v", params)

	return applicationRead(d, m, params)
}

// resourceApplicationRead read handler for resource
func resourceApplicationRead(d *schema.ResourceData, m interface{}) error {
	log.Debug("Lookup of ??? with id %s", d.Id()) // TBD

	return applicationRead(d, m, zabbix.Params{
		"applicationids": d.Id(),
	})
}

// applicationRead common application read function
func applicationRead(d *schema.ResourceData, m interface{}, params zabbix.Params) error {
	api := m.(*zabbix.API)

	log.Debug("Lookup of application with params %#v", params)

	apps, err := api.ApplicationsGet(params)
	if err != nil {
		return err
	}

	if len(apps) < 1 {
		d.SetId("")
		return nil
	}
	if len(apps) > 1 {
		return errors.New("multiple applications found")
	}
	app := apps[0]

	log.Debug("Got application: %+v", app)

	d.SetId(app.ApplicationID)
	d.Set("name", app.Name)
	d.Set("hostid", app.HostID)

	// templateSet := schema.NewSet(schema.HashString, []interface{}{})
	// for _, v := range app.ParentTemplateIDs {
	// 	templateSet.Add(v.TemplateID)
	// }
	// d.Set("templateids", templateSet)

	// flags : TBD ?

	return nil
}

// resourceApplicationUpdate terraform update resource handler
func resourceApplicationUpdate(d *schema.ResourceData, m interface{}) error {
	return errors.New("Unimplemented error")
	// api := m.(*zabbix.API)

	// item, err := buildApplicationObject(d)

	// if err != nil {
	// 	return err
	// }

	// item.ApplicationID = d.Id()

	// items := []zabbix.Application{*item}

	// err = api.ApplicationsUpdate(items)

	// if err != nil {
	// 	return err
	// }

	// return resourceApplicationRead(d, m)
}

// resourceApplicationDelete terraform delete resource handler
func resourceApplicationDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*zabbix.API)
	return api.ApplicationsDeleteByIds([]string{d.Id()})
}
