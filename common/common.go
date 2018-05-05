package common

import (
	"bufio"
	"io"
	"log"
	"net"
	"net/http"
	"os"

	"github.com/digitalocean/godo"
	"github.com/scaleway/go-scaleway"
	"golang.org/x/oauth2"
)

// TokenSource is now commented
type TokenSource struct {
	AccessToken string
}

// Token is now commented
func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{
		AccessToken: t.AccessToken,
	}
	return token, nil
}

var doPersonalAccessToken = os.Getenv("SCW_TOKEN")
var firewallID = os.Getenv("doFirewallID")

// FloatingIPAddress is the static IP for ackerson.de
var FloatingIPAddress = os.Getenv("doFloatingIP")

// PrepareScalewayLogin does what it says on the box
func PrepareScalewayLogin() *api.ScalewayAPI {
	api, err := api.NewScalewayAPI("my-organization", os.Getenv("SCW_TOKEN"), "dafuq", "")
	if err != nil {
		log.Printf("ERR: %s\n", err.Error())
	}
	return api
}

func prepareSSHipAddresses() []string {
	ipAddys := []string{os.Getenv("officeIP")}
	ipAddrs, _ := net.LookupIP(os.Getenv("homeDomain"))
	for _, ipAddr := range ipAddrs {
		ipAddys = append(ipAddys, ipAddr.String())
	}

	// https://docs.hetrixtools.com/uptime-monitoring-nodes/ (NYC,LON,FRA)
	//hetrixToolsCheckers := []string{"52.207.41.187", "51.140.35.64", "52.207.73.67", "52.23.120.125", "52.56.73.124", "139.162.228.62", "52.59.92.96", "78.46.88.58"}

	// switch to UptimeRobot
	uptimeRobotAddresses, err := urlToLines("https://uptimerobot.com/inc/files/ips/IPv4andIPv6.txt")
	if err != nil {
		log.Println(err.Error())
	}
	ipAddys = append(ipAddys, uptimeRobotAddresses...)

	return ipAddys
}

func urlToLines(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return linesFromReader(resp.Body)
}

func linesFromReader(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

// UpdateFirewall to maintain connectivity while Telekom rotates IPs
func UpdateFirewall() {
	//ipAddys := prepareSSHipAddresses()

	//client := PrepareScalewayLogin()
	//ctx := context.TODO()

	/*floatingIP, _, err := client.FloatingIPs.Get(ctx, os.Getenv("doFloatingIP"))
	for floatingIP.Droplet == nil {
		if err != nil {
			log.Println(err)
		}

		log.Println("floatIP not yet assigned...")
		time.Sleep(5 * time.Second)
		floatingIP, _, err = client.FloatingIPs.Get(ctx, os.Getenv("doFloatingIP"))
	}
	log.Println("update firewall for droplet: " + strconv.Itoa(floatingIP.Droplet.ID))

	updateRequest := &godo.FirewallRequest{
		Name: "SSH-HTTP-regulation",
		InboundRules: []godo.InboundRule{
			{
				Protocol:  "tcp",
				PortRange: "80",
				Sources: &godo.Sources{
					Addresses: []string{"0.0.0.0/0", "::/0"},
				},
			},
			{
				Protocol:  "tcp",
				PortRange: "443",
				Sources: &godo.Sources{
					Addresses: []string{"0.0.0.0/0", "::/0"},
				},
			},
			{
				Protocol:  "tcp",
				PortRange: "22",
				Sources: &godo.Sources{
					Addresses: ipAddys,
				},
			},
		},
		OutboundRules: []godo.OutboundRule{
			{
				Protocol:  "tcp",
				PortRange: "1-65535",
				Destinations: &godo.Destinations{
					Addresses: []string{"0.0.0.0/0", "::/0"},
				},
			},
			{
				Protocol: "icmp",
				Destinations: &godo.Destinations{
					Addresses: []string{"0.0.0.0/0", "::/0"},
				},
			},
			{
				Protocol:  "udp",
				PortRange: "1-65535",
				Destinations: &godo.Destinations{
					Addresses: []string{"0.0.0.0/0", "::/0"},
				},
			},
		},
		DropletIDs: []int{floatingIP.Droplet.ID},
	}*/

	/*firewallResp, _, err := client.Firewalls.Update(ctx, firewallID, updateRequest)
	if err == nil {
		log.Println(firewallResp)
	} else {
		log.Println(err)
	}*/
}

// DropletList does what it says on the box
func DropletList(client *godo.Client) ([]godo.Droplet, error) {
	list := []godo.Droplet{}

	// create options. initially, these will be blank
	opt := &godo.ListOptions{}
	for {
		droplets, resp, err := client.Droplets.List(oauth2.NoContext, opt)
		if err != nil {
			return nil, err
		}

		// append the current page's droplets to our list
		for _, d := range droplets {
			list = append(list, d)
		}

		// if we are at the last page, break out the for loop
		if resp.Links == nil || resp.Links.IsLastPage() {
			break
		}

		page, err := resp.Links.CurrentPage()
		if err != nil {
			return nil, err
		}

		// set the page we want for the next request
		opt.Page = page + 1
	}

	return list, nil
}

// DeleteDODroplet more here https://developers.digitalocean.com/documentation/v2/#delete-a-droplet
func DeleteDODroplet(ID int) string {
	var result string

	client := PrepareScalewayLogin()
	client.CheckCredentials()

	/*_, err := client.Droplets.Delete(oauth2.NoContext, ID)
	if err == nil {
		result = "Successfully deleted Droplet `" + strconv.Itoa(ID) + "`"
	} else {
		result = err.Error()
	}*/

	return result
}
