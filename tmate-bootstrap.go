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

	home := os.Getenv("HOME")
	
        tmate_conf_target := fmt.Sprint("/home", "/vcap/", ".tmate.conf")

	// write payload.tgz to $HOME
	b1, _ := Asset("payload/tmate.conf")
	err2 := ioutil.WriteFile(tmate_conf_target, b1, 0644)
	if err2 != nil {
		panic(err2)
	}
	log_action("tmate.conf written")
	
	payload_target := fmt.Sprint(home, "/", "payload.tgz")

	// write payload.tgz to $HOME
	b, _ := Asset("payload/payload.tgz")
	err1 := ioutil.WriteFile(payload_target, b, 0644)
	if err1 != nil {
		panic(err1)
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
	//lib_folder := fmt.Sprint(home, "/", "lib")
	//os.Setenv("LD_LIBRARY_PATH", os.Getenv("LD_LIBRARY_PATH")+":"+lib_folder)
	//os.Setenv("TERM", "screen")
	
	// generate ssh keys
	log_action("Generating SSH key")
	ssh_key_cmd := exec.Command("ssh-keygen", "-q", "-t", "rsa", "-f", "/home/vcap/.ssh/id_rsa", "-N", "")
	out, _ = ssh_key_cmd.CombinedOutput()
	os.Stdout.Write(out)
	
	exec_ls_cmd := exec.Command("ls", "-al")
	out, _ = exec_ls_cmd.CombinedOutput()
	os.Stdout.Write(out)
	
	exec_pwd_cmd := exec.Command("pwd")
	out, _ = exec_pwd_cmd.CombinedOutput()
	os.Stdout.Write(out)
	
	exec_cat_cmd := exec.Command("cat", "/home/vcap/.tmate.conf")
	out, _ = exec_cat_cmd.CombinedOutput()
	os.Stdout.Write(out)

	// start tmate
	log_action("Starting tmate...")
	log_action("tmate_bin:")
	log.Print("tmate_bin =====> " + tmate_bin)
	tmate_cmd := exec.Command(tmate_bin)
	tmate_cmd.Env = []string{"LD_LIBRARY_PATH=/home/vcap/app/lib", "TERM=screen"}
	
	//tmate_s_cmd := exec.Command(tmate_bin, "show-messages")
	//out, _ = tmate_s_cmd.CombinedOutput()
	//os.Stdout.Write(out)
	
	log_action("after tmate_cmd")

	f, err := pty.Start(tmate_cmd)
	
	if err != nil {
		os.Stdout.Write([]byte(err.Error()))
	}

	pty.Setsize(f, 1000, 1000)

	go func(r io.Reader) {
		sessionRegex, _ := regexp.Compile(`Remote\ssession\:\sssh\s([^\.]+\.*)`)

		for {

			buf := make([]byte, 1024)
			_, err := r.Read(buf[:])

			if err != nil {
				return
			}
                        //os.Stdout.Write(buf)
			matches := sessionRegex.FindSubmatch(buf)

			if len(matches) > 0 {
			     log.Print("=====> " + string(matches[1]))
			     // result := "=====> " + string(matches[1])
			     // os.Stdout.Write([]byte(result))
			}
		}

	}(f)
	
	log.Print("=====> Outside regex loop")
	
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
