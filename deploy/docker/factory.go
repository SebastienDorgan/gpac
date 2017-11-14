package docker

//go:generate rice embed-go
import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/SebastienDorgan/gpac/system"

	"github.com/GeertJohan/go.rice"
	"github.com/juju/loggo"
)

var logger = loggo.GetLogger("deploy.docker")

//Install install docker an a remote host
func Install(ssh *system.SSHHelper, useSudo bool) error {
	f, err := ioutil.TempFile("/tmp", "install_docker")
	defer f.Close()
	defer os.Remove(f.Name())
	if err != nil {
		return err
	}
	box, _ := rice.FindBox("scripts")
	script, err := box.Bytes("install_docker.sh")
	f.Write(script)
	ssh.Scp(f.Name(), "/tmp/install_docker.sh")
	cmd := "chmod a+x /tmp/install_docker.sh"
	_, _, timeout, err := ssh.Run("chmod a+x /tmp/install_docker.sh", 10*time.Second, useSudo)
	if err != nil {
		return nil
	}
	if timeout {
		msg := fmt.Sprintf("Timeout occurred running %s", cmd)
		logger.Errorf(msg)
		return fmt.Errorf(msg)
	}

	//Asynchronous call to /tmp/install_docker.sh
	stdoutChan, stderrChan, doneChan, errChan, err := ssh.Stream(cmd, 5*time.Minute, useSudo)
	if err != nil {
		return err
	}
	isTimeout := true
	err = nil
	for {
		select {
		case isTimeout = <-doneChan:
		case outline := <-stdoutChan:
			logger.Tracef(outline)
		case errline := <-stderrChan:
			logger.Errorf(errline)
		case err = <-errChan:
		}
		if isTimeout || err != nil {
			break
		}
	}

	// get exit code or command error.
	if err != nil {
		return err
	}

	// command time out
	if !isTimeout {
		msg := fmt.Sprintf("Timeout occurred running %s", cmd)
		logger.Errorf(msg)
		return fmt.Errorf(msg)
	}
	return nil
}
