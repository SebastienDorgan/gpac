package docker

import (
	"github.com/GeertJohan/go.rice/embedded"
	"time"
)

func init() {

	// define files
	file2 := &embedded.EmbeddedFile{
		Filename:    "install_docker.sh",
		FileModTime: time.Unix(1510382568, 0),
		Content:     string("#!/bin/bash\n\nexport DEBIAN_FRONTEND=noninteractive\napt-get install -qqy \\\n    apt-transport-https \\\n    ca-certificates \\\n    curl \\\n    software-properties-common\ncurl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add -\nadd-apt-repository \\\n   \"deb [arch=amd64] https://download.docker.com/linux/ubuntu \\\n   $(lsb_release -cs) \\\n   stable\"\napt-get update && apt-get install -qqy  docker-ce\n"),
	}

	// define dirs
	dir1 := &embedded.EmbeddedDir{
		Filename:   "",
		DirModTime: time.Unix(1510381157, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file2, // "install_docker.sh"

		},
	}

	// link ChildDirs
	dir1.ChildDirs = []*embedded.EmbeddedDir{}

	// register embeddedBox
	embedded.RegisterEmbeddedBox(`scripts`, &embedded.EmbeddedBox{
		Name: `scripts`,
		Time: time.Unix(1510381157, 0),
		Dirs: map[string]*embedded.EmbeddedDir{
			"": dir1,
		},
		Files: map[string]*embedded.EmbeddedFile{
			"install_docker.sh": file2,
		},
	})
}
