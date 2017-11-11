#!/bin/bash
ETCD_INTERFACE=$1
ETCD_DICOVERY_ID=$2

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

echo "start etcd discovery service"

cat <<- EOF > /etc/systemd/system/etcd_discovery.service
	[Unit]
	Description=etcd discovery service
	Documentation=https://github.com/coreos/etcd
	After=network.target

	[Service]
	Type=notify
    Restart=always
    RestartSec=5s
    LimitNOFILE=40000
    TimeoutStartSec=2min
	Environment="ETCD_DATA_DIR=/var/lib/etcd-discovery"
	Environment="ETCD_LISTEN_PEER_URLS=http://${ETCD_INTERFACE}:4001"
	Environment="ETCD_LISTEN_CLIENT_URLS=http://${ETCD_INTERFACE}:4000"
	Environment="ETCD_ADVERTISE_CLIENT_URLS=http://${ETCD_INTERFACE}:4000"
	ExecStart=/usr/local/bin/etcd -name etcd_discovery

	[Install]
	WantedBy=multi-user.target
EOF

systemctl enable etcd_discovery
systemctl start etcd_discovery

echo "register discovery key"
curl -L -X PUT "http://${ETCD_INTERFACE}:4000/v2/keys/discovery/${ETCD_DICOVERY_ID}/_config/size" -d value="1"

echo "start etcd service"
mkdir -p /var/lib/etcd
cat <<- EOF > /etc/systemd/system/etcd.service
	[Unit]
	Description=etcd key-value store
	Documentation=https://github.com/coreos/etcd
	After=etcd_discovery.service

	[Service]
	Type=notify
	Environment="ETCD_DATA_DIR=/var/lib/etcd"
	Environment="ETCD_DISCOVERY=http://${ETCD_INTERFACE}:4000/v2/keys/discovery/${ETCD_DICOVERY_ID}"
	Environment="ETCD_LISTEN_PEER_URLS=http://${ETCD_INTERFACE}:2380"
	Environment="ETCD_LISTEN_CLIENT_URLS=http://${ETCD_INTERFACE}:2379,http://127.0.0.1:2379"
	Environment="ETCD_INITIAL_ADVERTISE_PEER_URLS=http://${ETCD_INTERFACE}:2380"
	Environment="ETCD_ADVERTISE_CLIENT_URLS=http://${ETCD_INTERFACE}:2379"
	ExecStart=/usr/local/bin/etcd -name ${HOSTNAME}
	Restart=always
	RestartSec=10s
	LimitNOFILE=40000

	[Install]
	WantedBy=multi-user.target
EOF

systemctl enable etcd
systemctl start etcd
