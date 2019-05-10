package genymotion

import (
	"bufio"
	"fmt"
	"os/exec"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceGenymotion() *schema.Resource {
	return &schema.Resource{
		Create: resourceGenymotionCreate,
		Read:   resourceGenymotionRead,
		Delete: resourceGenymotionDelete,

		Schema: map[string]*schema.Schema{
			"recipe_uuid": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instance_uuid": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"adb_serial": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"adbconnect": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},
			"adb_serial_port": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceGenymotionCreate(d *schema.ResourceData, m interface{}) error {
	recipeUUID := d.Get("recipe_uuid").(string)
	name := d.Get("name").(string)
	adbSerialPort := d.Get("adb_serial_port").(string)
	connectedWithAdb := d.Get("adbconnect")

	// Start Genymotion Cloud Device
	cmd := exec.Command("gmsaas", "instances", "start", recipeUUID, name)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error: %s", output)
	}

	// Connect to adb with adb-serial-port
	if connectedWithAdb == true {
		uuid, _ := GetInstanceDetails(name)
		cmdArgs := []string{}
		if len(adbSerialPort) > 0 {
			cmdArgs = []string{"instances", "adbconnect", uuid, "--adb-serial-port", adbSerialPort}
		} else {
			cmdArgs = []string{"instances", "adbconnect", uuid}
		}

		cmd := exec.Command("gmsaas", cmdArgs...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("Error: %s", output)
		}
	}

	id := d.Get("name").(string)
	d.SetId(id)

	return resourceGenymotionRead(d, m)
}

func resourceGenymotionRead(d *schema.ResourceData, m interface{}) error {

	name := d.Get("name").(string)

	uuid, serial := GetInstanceDetails(name)
	d.Set("instance_uuid", uuid)
	d.Set("adb_serial", serial)

	return nil
}

func resourceGenymotionDelete(d *schema.ResourceData, m interface{}) error {
	name := d.Get("name").(string)

	uuid, _ := GetInstanceDetails(name)
	// Stop Genymotion Cloud Device
	cmd := exec.Command("gmsaas", "instances", "stop", uuid)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error: %s", output)
	}
	return nil
}

func GetInstanceDetails(name string) (string, string) {
	for index, line := range GetInstancesList() {
		if index >= 2 {
			s := strings.Split(line, "  ")
			if strings.Compare(s[1], name) == 0 {
				uuid := s[0]
				serial := s[2]
				return uuid, serial
			}
		}
	}
	return "", ""
}

func GetInstancesList() []string {
	adminList := exec.Command("gmsaas", "instances", "list")
	stdout, _ := adminList.StdoutPipe()
	adminList.Start()
	// Create new Scanner.
	scanner := bufio.NewScanner(stdout)
	result := []string{}
	// Use Scan.
	for scanner.Scan() {
		line := scanner.Text()
		// Append line to result.
		result = append(result, line)
	}
	return result
}
