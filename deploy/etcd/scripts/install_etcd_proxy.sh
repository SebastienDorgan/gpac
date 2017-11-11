#!/usr/bin/env bash
#!/bin/bash
ENDPOINTS=$1
PORT=${2:-23790}
echo "++++ Install etcd proxy"

cat <<- EOF > /etc/systemd/system/etcd-proxy.service
	[Unit]
	Description=etcd key-value store proxy
	Documentation=https://github.com/coreos/etcd
	After=network.target

	[Service]
	Type=notify
	Restart=always
	RestartSec=5s
	LimitNOFILE=40000
	TimeoutStartSec=2min
	ExecStart=/usr/local/bin/etcd grpc-proxy start --endpoints=${ENDPOINTS} --listen-addr=127.0.0.1:${PORT}

	[Install]
	WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable etcd-proxy
systemctl start etcd-proxy
