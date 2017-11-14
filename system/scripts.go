package system

import (
	"github.com/GeertJohan/go.rice"
)

//EmbeddedScript utility struct defining embedded scripts
type EmbeddedScript struct {
	box  *rice.Box
	name string
}

// func RunScript(ssh *SSHHelper, useSudo bool, name string, box *rice.Box, logger *loggo.Logger) (stdout string, err error) {
// 	f, err := ioutil.TempFile("/tmp", "install_docker")
// 	defer f.Close()
// 	defer os.Remove(f.Name())
// 	if err != nil {
// 		return "", err
// 	}
// 	script, err := box.Bytes("install_docker.sh")
// 	f.Write(script)
// 	ssh.Scp(f.Name(), "/tmp/install_docker.sh")
// 	cmd := "chmod a+x /tmp/install_docker.sh"
// 	_, _, timeout, err := ssh.Run("chmod a+x /tmp/install_docker.sh", 10*time.Second, useSudo)
// 	if err != nil {
// 		return "", nil
// 	}
// 	if timeout {
// 		msg := fmt.Sprintf("Timeout occurred running %s", cmd)
// 		logger.Errorf(msg)
// 		return fmt.Errorf(msg)
// 	}

// 	//Asynchronous call to /tmp/install_docker.sh
// 	stdoutChan, stderrChan, doneChan, errChan, err := ssh.Stream(cmd, 5*time.Minute, useSudo)
// 	if err != nil {
// 		return "", err
// 	}
// 	isTimeout := true
// 	err = nil
// 	for {
// 		select {
// 		case isTimeout = <-doneChan:
// 		case outline := <-stdoutChan:
// 			logger.Tracef(outline)
// 		case errline := <-stderrChan:
// 			logger.Errorf(errline)
// 		case err = <-errChan:
// 		}
// 		if isTimeout || err != nil {
// 			break
// 		}
// 	}

// 	// get exit code or command error.
// 	if err != nil {
// 		return "", err
// 	}

// 	// command time out
// 	if !isTimeout {
// 		msg := fmt.Sprintf("Timeout occurred running %s", cmd)
// 		logger.Errorf(msg)
// 		return fmt.Errorf(msg)
// 	}
// 	return "", nil
// }
