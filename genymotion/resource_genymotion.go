package genymotion

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func prepareGmsaasCommand(cmd *exec.Cmd) *exec.Cmd {
	cmd.Env = append(
		os.Environ(),
		"GMSAAS_USER_AGENT_EXTRA_DATA=Terraform",
	)
	return cmd
}

func resourceGenymotion() *schema.Resource {
	return &schema.Resource{
		Create: resourceGenymotionCreate,
		Read:   resourceGenymotionRead,
		Delete: resourceGenymotionDelete,

		Schema: map[string]*schema.Schema{
			"recipe_uuid": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instance_uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"adb_serial": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"adbconnect": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
				ForceNew: true,
			},
			"adb_serial_port": {
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
	cmd := prepareGmsaasCommand(exec.Command("gmsaas", "instances", "start", recipeUUID, name))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Error: %s", output)
	}

	// Connect to adb with adb-serial-port
	if connectedWithAdb == true {
		uuid, _ := GetInstanceDetails(name)
		if len(adbSerialPort) > 0 {
			cmd := prepareGmsaasCommand(exec.Command("gmsaas", "instances", "adbconnect", uuid, "--adb-serial-port", adbSerialPort))
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("Error: %s", output)
			}
		} else {
			cmd := prepareGmsaasCommand(exec.Command("gmsaas", "instances", "adbconnect", uuid))
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("Error: %s", output)
			}
		}
	}

	id := d.Get("name").(string)
	d.SetId(id)

	return resourceGenymotionRead(d, m)
}

func resourceGenymotionRead(d *schema.ResourceData, m interface{}) error {

	name := d.Get("name").(string)

	uuid, serial := GetInstanceDetails(name)
	if err := d.Set("instance_uuid", uuid); err != nil {
		return fmt.Errorf("Set instance_uuid failed, error: %s", err)
	}
	if err := d.Set("adb_serial", serial); err != nil {
		return fmt.Errorf("set adb_serial failed, error: %s", err)
	}

	return nil
}

func resourceGenymotionDelete(d *schema.ResourceData, m interface{}) error {
	name := d.Get("name").(string)

	uuid, _ := GetInstanceDetails(name)
	// Stop Genymotion Cloud Device
	cmd := prepareGmsaasCommand(exec.Command("gmsaas", "instances", "stop", uuid))
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
	adminList := prepareGmsaasCommand(exec.Command("gmsaas", "instances", "list"))
	stdout, _ := adminList.StdoutPipe()
	err := adminList.Start()
	if err != nil {
		log.Fatal(err)
	}
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
