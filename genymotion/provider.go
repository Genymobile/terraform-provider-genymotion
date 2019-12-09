package genymotion

import (
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("GENYMOTION_EMAIL", nil),
				Description: "Email for the Genymotion Cloud account",
			},

			"password": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("GENYMOTION_PASSWORD", nil),
				Description: "Password for the Genymotion Cloud account",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"genymotion_cloud": resourceGenymotion(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {

	var err error

	config := GenymotionConfig{
		Email:    d.Get("email").(string),
		Password: d.Get("password").(string),
	}

	// Connect to Genymotion account
	if err := config.connect(); err != nil {
		return nil, err
	}

	return nil, err
}

func (c GenymotionConfig) validate() error {
	var err *multierror.Error

	if c.Email == "" {
		err = multierror.Append(err, fmt.Errorf("Email must be configured for the Genymotion Cloud Provider"))
	}
	if c.Password == "" {
		err = multierror.Append(err, fmt.Errorf("Password must be configured for the Genymotion Cloud Provider"))
	}

	return err.ErrorOrNil()
}

func (c GenymotionConfig) connect() error {
	// Detect if user is already registered
	cmd := exec.Command(
		"gmsaas", "auth", "whoami")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("cmd.Run() failed with %s\n", err)
	}

	// If user is logged, do not login again
	if !validateEmail(strings.Trim(string(out), "\n")) {
		// Check mandatory fields
		if err := c.validate(); err != nil {
			return err
		}
		log.Println("[INFO] Login Genymotion Account")
		cmd := exec.Command(
			"gmsaas", "auth", "login", c.Email, c.Password)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("Error: %s", output)
		}
	} else {
		log.Println("[INFO] User is already logged")
	}
	return nil
}

func validateEmail(email string) bool {
	Re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	return Re.MatchString(email)
}

type GenymotionConfig struct {
	Email    string
	Password string
}
