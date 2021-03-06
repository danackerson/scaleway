package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/danackerson/scaleway/common"
	"github.com/digitalocean/godo"
	"github.com/scaleway/go-scaleway/types"
	"golang.org/x/oauth2"
)

var encodedDOSSHLoginPubKey = os.Getenv("encodedDOSSHLoginPubKey")
var circleCIBuild = os.Getenv("CIRCLE_BUILD_NUM")
var scwServerInfoSite = "https://cloud.scaleway.com/#/zones/par1/servers"

func main() {
	client := common.PrepareScalewayLogin()

	//client.CheckCredentials()
	//client.GetImages()
	//client.GetBootscripts()

	//linux4_15Kernel := "acbc1329-c6a0-4320-896b-56f9597f0c3c"
	//alpineLinux := "24141068-1043-4885-bf2b-8290f617e273"
	//dockerUbuntu := "c7dd3987-4b40-48d7-9481-b7fd2e597737"
	ubuntuBionicDocker := "00e48c6b-2896-4b2c-997e-aab7f4ab1c94"
	definition := &types.ScalewayServerDefinition{
		Name:           "test",
		CommercialType: "START1-S",
		Image:          &ubuntuBionicDocker,
		EnableIPV6:     true,
	}

	response, err := client.PostServer(*definition)
	if err != nil {
		log.Printf("CREATE ERR: %s", err.Error())
	}
	log.Printf(response)

	servers, err := client.GetServers(true, -1)
	if err != nil {
		log.Printf("GET ERR: %s\n", err.Error())
	}

	for _, server := range *servers {
		log.Printf("Created: %s\n", server.PublicAddress.IP)
		client.PostServerAction(server.Identifier, "poweron") // poweroff, poweron, reboot

		/*client.DeleteServer(server.Identifier)
		for _, volume := range server.Volumes {
			client.DeleteVolume(volume.Identifier)
		}*/
	}
	//log.Printf("%v", servers)

	return
	/*
		fnPtr := flag.String("fn", "updateDNS|createNewServer", "which function to run")
		dropletIDPtr := flag.String("dropletID", "<digitalOceanDropletID>", "DO droplet to attach floatingIP to")
		flag.Parse()
		if *fnPtr == "createNewServer" {
			droplet := createDroplet(client)
			waitUntilDropletReady(client, droplet.ID)
			droplet, _, _ = client.Droplets.Get(oauth2.NoContext, droplet.ID)

			ipv4, _ := droplet.PublicIPv4()
			addr := scwServerInfoSite + strconv.Itoa(droplet.ID)
			fmt.Printf("%s: %s @%s\n", ipv4, droplet.Name, addr)

			// Write /tmp/new_digital_ocean_droplet_params
			envVarsFile := []byte("export NEW_SERVER_IPV4=" + ipv4 + "\nexport NEW_DROPLET_ID=" + strconv.Itoa(droplet.ID) + "\n")
			err := ioutil.WriteFile("/tmp/new_digital_ocean_droplet_params", envVarsFile, 0644)
			if err != nil {
				fmt.Printf("Failed to write /tmp/new_digital_ocean_droplet_params: %s", err)
			}

			var firewallID = os.Getenv("doFirewallID")
			_, err2 := client.Firewalls.AddDroplets(oauth2.NoContext, firewallID, droplet.ID)
			if err2 != nil {
				fmt.Printf("Failed to add droplet to Firewall: %s", err2)
			}
		} else {
			dropletID, _ := strconv.Atoi(*dropletIDPtr)
			droplet, _, _ := client.Droplets.Get(oauth2.NoContext, dropletID)
			fmt.Printf("\ngoing to work on DropletID: %d\n", droplet.ID)

			reassignFloatingIP(client, droplet)

			// update ipv6 DNS entry to new droplet
			ipv6, _ := droplet.PublicIPv6()
			fmt.Printf("new IPv6 addr: %s\n", ipv6)
			updateIPV6(client, ipv6, "ackerson.de", 23738236)
			updateIPV6(client, ipv6, "battlefleet.online", 30208348)

			common.UpdateFirewall()
		}*/
}

func updateIPV6(client *godo.Client, ipv6 string, hostname string, domainID int) {
	record, _, _ := client.Domains.Record(oauth2.NoContext, hostname, domainID)
	fmt.Printf("current IPv6 %s: %s => %s", record.Name, record.Type, record.Data)

	editRequest := &godo.DomainRecordEditRequest{
		Type: record.Type,
		Name: record.Name,
		Data: strings.ToLower(ipv6),
	}
	_, _, err := client.Domains.EditRecord(oauth2.NoContext, hostname, domainID, editRequest)
	for err != nil {
		fmt.Printf("FAIL domain update IPv6: %s\n", err)
		time.Sleep(5 * time.Second)
		_, _, err = client.Domains.EditRecord(oauth2.NoContext, hostname, domainID, editRequest)
	}
}

// wait for in-progress actions to complete
func waitUntilDropletReady(client *godo.Client, dropletID int) {
	opt := &godo.ListOptions{}

	for ready := false; !ready; {
		actions, _, _ := client.Droplets.Actions(oauth2.NoContext, dropletID, opt)
		ready = true
		for j, action := range actions {
			fmt.Printf("%d: %s => %s\n", j, action.Type, action.Status)
			if action.Status == "in-progress" {
				ready = false
			}
		}
		if !ready {
			time.Sleep(time.Second * 5)
		}
	}
}

func reassignFloatingIP(client *godo.Client, droplet *godo.Droplet) {
	client.FloatingIPActions.Unassign(oauth2.NoContext, common.FloatingIPAddress)

	_, _, err := client.FloatingIPActions.Assign(oauth2.NoContext, common.FloatingIPAddress, droplet.ID)
	for err != nil {
		fmt.Printf("WARN: %s\n", err.Error())
		time.Sleep(5 * time.Second)
		_, _, err = client.FloatingIPActions.Assign(oauth2.NoContext, common.FloatingIPAddress, droplet.ID)
	}
}

func createDroplet(client *godo.Client) *godo.Droplet {
	var newDroplet *godo.Droplet

	fingerprint := "e0:a3:4c:5a:5a:1b:9c:bb:b5:51:a7:7f:62:27:51:96"
	dropletName := "b" + circleCIBuild + ".ackerson.de"

	sshKeys := []godo.DropletCreateSSHKey{}
	sshKeys = append(sshKeys, godo.DropletCreateSSHKey{Fingerprint: fingerprint})

	digitaloceanIgnitionJSON, err := ioutil.ReadFile("digitalocean_ignition.json")
	if err != nil {
		fmt.Printf("Failed to read JSON file: %s", err)
	} else {
		createRequest := &godo.DropletCreateRequest{
			Name:   dropletName,
			Region: "fra1",
			Size:   "s-1vcpu-1gb",
			Image: godo.DropletCreateImage{
				Slug: "coreos-stable",
			},
			IPv6:     true,
			SSHKeys:  sshKeys,
			UserData: string(digitaloceanIgnitionJSON),
		}

		newDroplet, _, err = client.Droplets.Create(oauth2.NoContext, createRequest)
		if err != nil {
			fmt.Printf("\nUnexpected ERROR: %s\n\n", err)
			os.Exit(1)
		}
	}

	return newDroplet
}
