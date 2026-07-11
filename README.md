# Sericulture Backend

This repository contains the backend service for the Sericulture application.

## Service Commands

Use the following `systemctl` commands to control the service:

- Start: `sudo systemctl start sericulture`
- Stop: `sudo systemctl stop sericulture`
- Restart: `sudo systemctl restart sericulture`
- Enable at boot: `sudo systemctl enable sericulture`
- Check status: `sudo systemctl status sericulture`

## Logs

To view live log output, run:

```bash
tail -f logs/output.log
```

## Service Unit File

The systemd unit file is located at:

```text
/etc/systemd/system/sericulture.service
```

## Notes

- Make sure the service is installed and the unit file is present before using `systemctl`.
- If the service does not start, inspect the journal logs with `sudo journalctl -u sericulture.service`.

## Export log
cd /path/to/logs
python3 -m http.server 4000

## Kill Port
sudo fuser -k 3000/tcp