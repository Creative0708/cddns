package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"

	"github.com/manifoldco/promptui"
)

func main() {
	log.Default().SetFlags(0)

	setup := flag.Bool("setup", false, "run setup")
	ipOverride := flag.String("i", "", "override ip address to update with")

	flag.Parse()

	config, err := loadConfig()
	if os.IsNotExist(err) {
		fmt.Println("no config file found. entering setup")
		*setup = true
	}
	if *setup {
		config, err = handleSetup()
		if err != nil {
			log.Fatal("failed to setup: ", err)
		}
	}

	if err != nil {
		log.Fatal("cannot load config: ", err)
	}

	serverLoc := config.ServerLocation
	if !strings.HasSuffix(serverLoc, "/") {
		serverLoc += "/"
	}

	// TODO: configurable ip providers
	var ipProvider IpProvider = IpifyIpProvider{}

	var ip net.IP
	if *ipOverride != "" {
		ip = net.ParseIP(*ipOverride)
		if ip == nil {
			log.Fatal("invalid ip address: ", *ipOverride)
		}
	} else {
		ip, err = ipProvider.GetIp(http.DefaultClient)
		if err != nil {
			log.Fatal("failed to get IP address: ", err)
		}
	}

	req, err := http.NewRequest("POST", serverLoc+"v1/update", strings.NewReader(ip.String()))
	if err != nil {
		log.Fatal("failed to construct request: ", err)
	}
	req.Header.Add("Authorization", "Bearer "+config.ApiKey)

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		log.Fatal("failed to post to server: ", err)
	}
	if resp.Status[0] != '2' {
		log.Fatal("invalid server response: ", resp.Status)
	}
	fmt.Println("successfully updated ip address to", ip)
}

type UrlError string

func (e UrlError) Error() string {
	return string(e)
}

func handleSetup() (*Config, error) {
	prompt := promptui.Prompt{
		Label:   "server URL",
		Default: "https://example.com/cddns",
		Validate: func(s string) error {
			url, err := url.Parse(s)
			if err != nil {
				return err
			}
			if url.Hostname() == "example.com" {
				return UrlError("cannot use example host")
			}
			if url.Scheme != "https" && url.Scheme != "http" {
				return UrlError("url must use HTTP(S)")
			}
			return err
		},
	}
	server_url, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	prompt = promptui.Prompt{
		Label:    "API key",
		Default:  APIKEY_PREFIX + "********",
		Validate: validateApiKey,
	}

	api_key, err := prompt.Run()
	if err != nil {
		return nil, err
	}

	config := Config{
		ServerLocation: server_url,
		ApiKey:         api_key,
	}

	saveConfig(&config)

	if runtime.GOOS == "linux" {

		prompt := promptui.Select{
			Label: "run automatically?",
			Items: []string{
				"on reboot",
				"hourly",
				"daily",
				"weekly",
				"don't run automatically",
			},
		}
		idx, _, err := prompt.Run()
		if err != nil {
			return nil, err
		}
		if idx < INSTALL_TYPE_LENGTH {
			err := installSystemdService(getPermanentBinaryPath(), InstallType(idx))
			if err != nil {
				log.Print("warning: failed to install systemd service: ", err)
			}
		}
	}

	return &config, nil
}
