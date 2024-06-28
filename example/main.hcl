
file "/tmp/blah" "neat" {
	content = "ahoy there"
}

exec {
	cmd = "/usr/bin/cat ${file.neat.path}"
	after = file.neat
}

file "/tmp/b" {
	content = "balls"
}

service "ahoy" {
	running = true
	enabled = true
}

file "/lib/systemd/system/ahoy.service" {
	after = service.ahoy
	template = j2
	content = <<EOF
[Unit]
Description=Ahoy matey
After=network-online.target network.target
Wants=network-online.target

[Service]
Type=simple
User=ahoy
Group=ahoy
ExecStart=/usr/bin/echo ahoy

[Install]
WantedBy=multi-user.target
EOF
	onchange = service.ahoy.restart
}
