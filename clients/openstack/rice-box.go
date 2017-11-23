package openstack

import (
	"github.com/GeertJohan/go.rice/embedded"
	"time"
)

func init() {

	// define files
	file2 := &embedded.EmbeddedFile{
		Filename:    "userdata.sh",
		FileModTime: time.Unix(1511385909, 0),
		Content:     string("#!/bin/bash\n\nadduser {{.User}} -gecos \"\" --disabled-password\necho \"{{.User}} ALL=(ALL) NOPASSWD:ALL\" >> /etc/sudoers\n\nmkdir /home/{{.User}}/.ssh\necho \"{{.Key}}\" > /home/{{.User}}/.ssh/authorized_keys"),
	}

	// define dirs
	dir1 := &embedded.EmbeddedDir{
		Filename:   "",
		DirModTime: time.Unix(1511346294, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file2, // "userdata.sh"

		},
	}

	// link ChildDirs
	dir1.ChildDirs = []*embedded.EmbeddedDir{}

	// register embeddedBox
	embedded.RegisterEmbeddedBox(`scripts`, &embedded.EmbeddedBox{
		Name: `scripts`,
		Time: time.Unix(1511346294, 0),
		Dirs: map[string]*embedded.EmbeddedDir{
			"": dir1,
		},
		Files: map[string]*embedded.EmbeddedFile{
			"userdata.sh": file2,
		},
	})
}
