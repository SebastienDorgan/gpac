#!/usr/bin/env bash
#!/bin/bash
LEADER_ETCD_INTERFACE=$1
ETCD_INTERFACE=$2
ETCD_DICOVERY_ID=$3

output=$(mktemp -d)
cd "${output}"


ETCD_VER=v3.2.2
DOWNLOAD_URL=https://github.com/coreos/etcd/releases/download
curl -L "${DOWNLOAD_URL}/${ETCD_VER}/etcd-${ETCD_VER}-linux-amd64.tar.gz" -o "${output}/etcd-${ETCD_VER}-linux-amd64.tar.gz"
mkdir -p "${output}/etcd"
tar xzvf "${output}/etcd-${ETCD_VER}-linux-amd64.tar.gz" -C "${output}/etcd" --strip-components=1


chmod a+x  "${output}/etcd/etcd"
chmod a+x  "${output}/etcd/etcdctl"
cp -f "${output}/etcd/etcd" "/usr/local/bin"
cp -f "${output}/etcd/etcdctl" "/usr/local/bin"

rm -rf "${output}"

# cloudwatt persistence bug turn around
rm -rf /var/lib/etcd
echo "start etcd service"
mkdir -p /var/lib/etcd
cat <<- EOF > /etc/systemd/system/etcd.service
	[Unit]
	Description=etcd key-value store
	Documentation=https://github.com/coreos/etcd
	After=network.target

	[Service]
	Type=notify
	Restart=always
	RestartSec=5s
	LimitNOFILE=40000
	TimeoutStartSec=2min
	Environment="ETCD_DATA_DIR=/var/lib/etcd"
	Environment="ETCD_DISCOVERY=http://${LEADER_ETCD_INTERFACE}:4000/v2/keys/discovery/${ETCD_DICOVERY_ID}"
	Environment="ETCD_LISTEN_PEER_URLS=http://${ETCD_INTERFACE}:2380"
	Environment="ETCD_LISTEN_CLIENT_URLS=http://${ETCD_INTERFACE}:2379,http://127.0.0.1:2379"
	Environment="ETCD_ADVERTISE_CLIENT_URLS=http://${ETCD_INTERFACE}:2379"
	ExecStart=/usr/local/bin/etcd -name ${HOSTNAME}


	[Install]
	WantedBy=multi-user.target
EOF

systemctl enable etcd
systemctl start etcd
