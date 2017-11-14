package etcd

import (
	"time"

	"github.com/GeertJohan/go.rice/embedded"
)

func init() {

	// define files
	file2 := &embedded.EmbeddedFile{
		Filename:    "install_etcd.sh",
		FileModTime: time.Unix(1510382680, 0),
		Content:     string("#!/usr/bin/env bash\n#!/bin/bash\nLEADER_ETCD_INTERFACE=$1\nETCD_INTERFACE=$2\nETCD_DICOVERY_ID=$3\n\noutput=$(mktemp -d)\ncd \"${output}\"\n\n\nETCD_VER=v3.2.2\nDOWNLOAD_URL=https://github.com/coreos/etcd/releases/download\ncurl -L \"${DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz\" -o \"${output}/etcd-${ETCD_VER}-linux-amd64.tar.gz\"\nmkdir -p \"${output}/etcd\"\ntar xzvf \"${output}/etcd-${ETCD_VER}-linux-amd64.tar.gz\" -C \"${output}/etcd\" --strip-components=1\n\n\nchmod a+x  \"${output}/etcd/etcd\"\nchmod a+x  \"${output}/etcd/etcdctl\"\ncp -f \"${output}/etcd/etcd\" \"/usr/local/bin\"\ncp -f \"${output}/etcd/etcdctl\" \"/usr/local/bin\"\n\nrm -rf \"${output}\"\n\n# cloudwatt persistence bug turn around\nrm -rf /var/lib/etcd\necho \"start etcd service\"\nmkdir -p /var/lib/etcd\ncat <<- EOF > /etc/systemd/system/etcd.service\n\t[Unit]\n\tDescription=etcd key-value store\n\tDocumentation=https://github.com/coreos/etcd\n\tAfter=network.target\n\n\t[Service]\n\tType=notify\n\tRestart=always\n\tRestartSec=5s\n\tLimitNOFILE=40000\n\tTimeoutStartSec=2min\n\tEnvironment=\"ETCD_DATA_DIR=/var/lib/etcd\"\n\tEnvironment=\"ETCD_DISCOVERY=http://${LEADER_ETCD_INTERFACE}:4000/v2/keys/discovery/${ETCD_DICOVERY_ID}\"\n\tEnvironment=\"ETCD_LISTEN_PEER_URLS=http://${ETCD_INTERFACE}:2380\"\n\tEnvironment=\"ETCD_LISTEN_CLIENT_URLS=http://${ETCD_INTERFACE}:2379,http://127.0.0.1:2379\"\n\tEnvironment=\"ETCD_ADVERTISE_CLIENT_URLS=http://${ETCD_INTERFACE}:2379\"\n\tExecStart=/usr/local/bin/etcd -name ${HOSTNAME}\n\n\n\t[Install]\n\tWantedBy=multi-user.target\nEOF\n\nsystemctl enable etcd\nsystemctl start etcd\n"),
	}
	file3 := &embedded.EmbeddedFile{
		Filename:    "install_etcd_discovery_and_etcd.sh",
		FileModTime: time.Unix(1510383046, 0),
		Content:     string("#!/bin/bash\nETCD_INTERFACE=$1\nETCD_DICOVERY_ID=$2\n\noutput=$(mktemp -d)\ncd \"${output}\"\n\n\nETCD_VER=v3.2.2\nDOWNLOAD_URL=https://github.com/coreos/etcd/releases/download\ncurl -L \"${DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz\" -o \"${output}/etcd-${ETCD_VER}-linux-amd64.tar.gz\"\nmkdir -p \"${output}/etcd\"\ntar xzvf \"${output}/etcd-${ETCD_VER}-linux-amd64.tar.gz\" -C \"${output}/etcd\" --strip-components=1\n\n\nchmod a+x  \"${output}/etcd/etcd\"\nchmod a+x  \"${output}/etcd/etcdctl\"\ncp -f \"${output}/etcd/etcd\" \"/usr/local/bin\"\ncp -f \"${output}/etcd/etcdctl\" \"/usr/local/bin\"\n\nrm -rf \"${output}\"\n\necho \"start etcd discovery service\"\n\ncat <<- EOF > /etc/systemd/system/etcd_discovery.service\n\t[Unit]\n\tDescription=etcd discovery service\n\tDocumentation=https://github.com/coreos/etcd\n\tAfter=network.target\n\n\t[Service]\n\tType=notify\n    Restart=always\n    RestartSec=5s\n    LimitNOFILE=40000\n    TimeoutStartSec=2min\n\tEnvironment=\"ETCD_DATA_DIR=/var/lib/etcd-discovery\"\n\tEnvironment=\"ETCD_LISTEN_PEER_URLS=http://${ETCD_INTERFACE}:4001\"\n\tEnvironment=\"ETCD_LISTEN_CLIENT_URLS=http://${ETCD_INTERFACE}:4000\"\n\tEnvironment=\"ETCD_ADVERTISE_CLIENT_URLS=http://${ETCD_INTERFACE}:4000\"\n\tExecStart=/usr/local/bin/etcd -name etcd_discovery\n\n\t[Install]\n\tWantedBy=multi-user.target\nEOF\n\nsystemctl enable etcd_discovery\nsystemctl start etcd_discovery\n\necho \"register discovery key\"\ncurl -L -X PUT \"http://${ETCD_INTERFACE}:4000/v2/keys/discovery/${ETCD_DICOVERY_ID}/_config/size\" -d value=\"1\"\n\necho \"start etcd service\"\nmkdir -p /var/lib/etcd\ncat <<- EOF > /etc/systemd/system/etcd.service\n\t[Unit]\n\tDescription=etcd key-value store\n\tDocumentation=https://github.com/coreos/etcd\n\tAfter=etcd_discovery.service\n\n\t[Service]\n\tType=notify\n\tEnvironment=\"ETCD_DATA_DIR=/var/lib/etcd\"\n\tEnvironment=\"ETCD_DISCOVERY=http://${ETCD_INTERFACE}:4000/v2/keys/discovery/${ETCD_DICOVERY_ID}\"\n\tEnvironment=\"ETCD_LISTEN_PEER_URLS=http://${ETCD_INTERFACE}:2380\"\n\tEnvironment=\"ETCD_LISTEN_CLIENT_URLS=http://${ETCD_INTERFACE}:2379,http://127.0.0.1:2379\"\n\tEnvironment=\"ETCD_INITIAL_ADVERTISE_PEER_URLS=http://${ETCD_INTERFACE}:2380\"\n\tEnvironment=\"ETCD_ADVERTISE_CLIENT_URLS=http://${ETCD_INTERFACE}:2379\"\n\tExecStart=/usr/local/bin/etcd -name ${HOSTNAME}\n\tRestart=always\n\tRestartSec=10s\n\tLimitNOFILE=40000\n\n\t[Install]\n\tWantedBy=multi-user.target\nEOF\n\nsystemctl enable etcd\nsystemctl start etcd\n"),
	}
	file4 := &embedded.EmbeddedFile{
		Filename:    "install_etcd_proxy.sh",
		FileModTime: time.Unix(1510353708, 0),
		Content:     string("#!/usr/bin/env bash\n#!/bin/bash\nENDPOINTS=$1\nPORT=${2:-23790}\necho \"++++ Install etcd proxy\"\n\ncat <<- EOF > /etc/systemd/system/etcd-proxy.service\n\t[Unit]\n\tDescription=etcd key-value store proxy\n\tDocumentation=https://github.com/coreos/etcd\n\tAfter=network.target\n\n\t[Service]\n\tType=notify\n\tRestart=always\n\tRestartSec=5s\n\tLimitNOFILE=40000\n\tTimeoutStartSec=2min\n\tExecStart=/usr/local/bin/etcd grpc-proxy start --endpoints=${ENDPOINTS} --listen-addr=127.0.0.1:${PORT}\n\n\t[Install]\n\tWantedBy=multi-user.target\nEOF\n\nsystemctl daemon-reload\nsystemctl enable etcd-proxy\nsystemctl start etcd-proxy\n"),
	}
	file5 := &embedded.EmbeddedFile{
		Filename:    "install_etcd_worker.sh",
		FileModTime: time.Unix(1510353721, 0),
		Content:     string("#!/usr/bin/env bash\n#!/bin/bash\nLEADER_ETCD_INTERFACE=$1\nETCD_INTERFACE=$2\nETCD_DICOVERY_ID=$3\necho \"++++ Install etcdctl\"\n\noutput=$(mktemp -d)\ncd \"${output}\"\n\n\nETCD_VER=v3.2.2\nDOWNLOAD_URL=https://github.com/coreos/etcd/releases/download\ncurl -L \"${DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz\" -o \"${output}/etcd-${ETCD_VER}-linux-amd64.tar.gz\"\nmkdir -p \"${output}/etcd\"\ntar xzvf \"${output}/etcd-${ETCD_VER}-linux-amd64.tar.gz\" -C \"${output}/etcd\" --strip-components=1\n\n\nchmod a+x  \"${output}/etcd/etcd\"\nchmod a+x  \"${output}/etcd/etcdctl\"\ncp -f \"${output}/etcd/etcd\" \"/usr/local/bin\"\ncp -f \"${output}/etcd/etcdctl\" \"/usr/local/bin\"\n\nrm -rf \"${output}\"\n# cloudwatt persistence bug turn around\nrm -rf /var/lib/etcd\necho \"start etcd service\"\nmkdir -p /var/lib/etcd\ncat <<- EOF > /etc/systemd/system/etcd.service\n\t[Unit]\n\tDescription=etcd key-value store\n\tDocumentation=https://github.com/coreos/etcd\n\tAfter=network.target\n\n\t[Service]\n\tType=notify\n\tRestart=always\n\tRestartSec=5s\n\tLimitNOFILE=40000\n    TimeoutStartSec=2min\n\tEnvironment=\"ETCD_DATA_DIR=/var/lib/etcd\"\n\tEnvironment=\"ETCD_DISCOVERY=http://${LEADER_ETCD_INTERFACE}:4000/v2/keys/discovery/${ETCD_DICOVERY_ID}\"\n\tEnvironment=\"ETCD_LISTEN_CLIENT_URLS=http://${ETCD_INTERFACE}:2379,http://127.0.0.1:2379\"\n    ExecStartPre=-/bin/rm -Rf /var/lib/etcd\n    ExecStartPre=-/bin/mkdir /var/lib/etcd\n\tExecStart=/usr/local/bin/etcd -name ${HOSTNAME} --proxy on\n\n\n\t[Install]\n\tWantedBy=multi-user.target\nEOF\n\nsystemctl enable etcd\nsystemctl start etcd\n"),
	}

	// define dirs
	dir1 := &embedded.EmbeddedDir{
		Filename:   "",
		DirModTime: time.Unix(1510381142, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file2, // "install_etcd.sh"
			file3, // "install_etcd_discovery_and_etcd.sh"
			file4, // "install_etcd_proxy.sh"
			file5, // "install_etcd_worker.sh"

		},
	}

	// link ChildDirs
	dir1.ChildDirs = []*embedded.EmbeddedDir{}

	// register embeddedBox
	embedded.RegisterEmbeddedBox(`scripts`, &embedded.EmbeddedBox{
		Name: `scripts`,
		Time: time.Unix(1510381142, 0),
		Dirs: map[string]*embedded.EmbeddedDir{
			"": dir1,
		},
		Files: map[string]*embedded.EmbeddedFile{
			"install_etcd.sh":                    file2,
			"install_etcd_discovery_and_etcd.sh": file3,
			"install_etcd_proxy.sh":              file4,
			"install_etcd_worker.sh":             file5,
		},
	})
}
