# Pi Iteration Workflow — Recipes

## Host access table

| Alias | Host | Notes |
|---|---|---|
| `rpi-5-direct` | `192.168.250.2` | Pi 5, direct Ethernet — most reliable |
| `rpi-5-local` | `192.168.0.50` | Pi 5, WiFi |
| `rpi-5-remote` | via Cloudflare tunnel | often broken/unauthenticated, don't rely on it |
| `rpi-zero-ethernet` | `ralvarezdev-raspberrypi-zero.local` / `192.168.0.51` | Pi Zero, WiFi (name is historical — USB gadget isn't set up) |

Raw-IP fallback when mDNS resolution fails (raw IP doesn't match the `~/.ssh/config` `Host` pattern, so `-i` must be explicit):

```
ssh -i "D:\Secrets\ssh\raspberrypi\id_ed25519_ralvarezdev" ralvarezdev@192.168.0.51
```

## Regenerating pixi.lock

On the Pi 5:

```
ssh rpi-5-direct
cd ~/voldemorbot && git pull
cd platform/robot && ~/.pixi/bin/pixi lock
```

Bring the lockfile back to Windows, commit, push, then drop the now-stale local diff on the Pi 5:

```
scp rpi-5-direct:~/voldemorbot/platform/robot/pixi.lock .
git add platform/robot/pixi.lock && git commit -m "..." && git push
```

Then on the Pi 5:

```
git checkout -- platform/robot/pixi.lock && git pull
```

## Long-running commands over screen

```
ssh -t -i "D:\Secrets\ssh\raspberrypi\id_ed25519_ralvarezdev" ralvarezdev@192.168.0.51
screen -S work
<commands>
```

Detach with `Ctrl+A` `D` once the process is running — it keeps going on the Pi independent of the SSH connection. Reattach later with `screen -d -r work`.
