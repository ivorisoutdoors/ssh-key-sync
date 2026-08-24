# ssh-key-sync

Sync ssh keys from GitHub (or Codeberg) to authorized_keys.

## Install

Grab the latest release and put it in `/usr/local/bin` (or somewhere else in PATH).

### systemd

```systemd
[Unit]
Description=Sync SSH authorized keys
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
Restart=always
RestartSec=10
User=<user>
ExecStart=/usr/local/bin/ssh-key-sync -user <gh-username> -interval 1h [-verbose]

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl start ssh-key-sync
sudo systemctl enable ssh-key-sync.service
```

### MacOS

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
    <dict>
        <key>Label</key>
        <string>com.ivorisoutdoors.ssh-key-sync</string>
        <key>ProgramArguments</key>
        <array>
            <string>/usr/local/bin/ssh-key-sync</string>
            <string>-user</string>
            <string><gh-username></string>
            <string>-interval</string>
            <string>1h</string>
            <string>-verbose</string>
        </array>
        <key>RunAtLoad</key>
        <true/>
        <key>KeepAlive</key>
        <true/>
        <key>StandardErrorPath</key>
        <string>/opt/homebrew/var/log/ssh-key-sync.log</string>
        <key>StandardOutPath</key>
        <string>/opt/homebrew/var/log/ssh-key-sync.log</string>
    </dict>
</plist>
```

## Codeberg

You can pull keys from Codeberg as well by passing `-url https://codeberg.org`.
