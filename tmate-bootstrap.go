package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"regexp"

	"github.com/danhigham/pty"
)

func log_action(text string) {
	os.Stdout.Write([]byte("-----> " + text + "\n"))
}

func main() {

	log_action("Extracting payload...")
	
	set tmate-server-host "10.5.185.177"
	set tmate-server-port 222
	set tmate-server-dsa-fingerprint   "e2:99:f9:c8:2f:47:88:8f:4a:49:45:77:b7:2a:a6:23"
	set tmate-server-rsa-fingerprint   "1d:48:87:4d:38:f0:72:ce:53:29:04:bf:a1:c4:19:04"
	set tmate-server-ecdsa-fingerprint "67:ca:ea:88:7d:7b:18:a1:09:27:00:76:dd:7f:3d:a9"
	set tmate-identity ""              


	home := os.Getenv("HOME")
	payload_target := fmt.Sprint(home, "/", "payload.tgz")

	// write payload.tgz to $HOME
	b, _ := Asset("payload/payload.tgz")
	err := ioutil.WriteFile(payload_target, b, 0644)
	if err != nil {
		panic(err)
	}

	// decompress payload.tgz
	unzip_cmd := exec.Command("tar", "xvzf", payload_target, "-C", home)
	out, err := unzip_cmd.CombinedOutput()

	os.Stdout.Write(out)

	// exec child process
	if len(os.Args) > 1 {
		go func() {
			log_action("Running child process")

			child_bin, args := os.Args[1], os.Args[2:len(os.Args)]
			child_cmd := exec.Command(child_bin, args...)
			out, err = child_cmd.CombinedOutput()
			os.Stdout.Write(out)

			if err != nil {
				os.Stdout.Write([]byte(err.Error()))
			}
		}()
	}

	// give tmate +x permissions
	log_action("Fixing permissions")
	tmate_bin := fmt.Sprint(home, "/", "bin", "/", "tmate")
	exec_perm_cmd := exec.Command("chmod", "+x", tmate_bin)
	out, _ = exec_perm_cmd.CombinedOutput()
	os.Stdout.Write(out)

	// add lib folder to LD_LIBRARY_PATH
	log_action("Setting env")
	lib_folder := fmt.Sprint(home, "/", "lib")
	os.Setenv("LD_LIBRARY_PATH", os.Getenv("LD_LIBRARY_PATH")+":"+lib_folder)
	os.Setenv("TERM", "screen-256color")

	// generate ssh keys
	log_action("Generating SSH key")
	ssh_key_cmd := exec.Command("ssh-keygen", "-q", "-t", "rsa", "-f", "/home/vcap/.ssh/id_rsa", "-N", "")
	out, _ = ssh_key_cmd.CombinedOutput()
	os.Stdout.Write(out)

	// start tmate
	log_action("Starting tmate...")
	log_action("tmate_bin:")
	log.Print("tmate_bin =====> " + tmate_bin)
	tmate_cmd := exec.Command(tmate_bin)
	
	os.Stdout.Write(tmate_cmd)
	f, err := pty.Start(tmate_cmd)

	if err != nil {
		os.Stdout.Write([]byte(err.Error()))
	}

	pty.Setsize(f, 1000, 1000)

	go func(r io.Reader) {
		sessionRegex, _ := regexp.Compile(`Remote\ssession\:\sssh\s([^\.]+\.tmate.io)`)

		for {

			buf := make([]byte, 1024)
			_, err := r.Read(buf[:])

			if err != nil {
				return
			}
                        os.Stdout.Write(buf)
			matches := sessionRegex.FindSubmatch(buf)

			if len(matches) > 0 {
				log.Print("=====> " + string(matches[1]))
				// result := "=====> " + string(matches[1])
				// os.Stdout.Write([]byte(result))
			} else {
				log.Print("No Match =====> ")	
			}
		}

	}(f)

	// set up a reverse proxy
	serverUrl, _ := url.Parse("http://127.0.0.1:8080")
	reverseProxy := httputil.NewSingleHostReverseProxy(serverUrl)

	http.Handle("/", reverseProxy)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	l, _ := net.Listen("tcp", ":"+os.Getenv("PORT"))

	go func() {
		for _ = range c {

			// sig is a ^C, handle it
			log.Print("Stopping tmate...")

			l.Close()
			f.Close()
		}
	}()

	http.Serve(l, nil)
	log.Print(err)
}
