# s3Sync
Sync S3 files from a remote server to a local S3 server (like MiniO)

# Sync files from one MiniO to another MiniO with CRON tab

# See logs from container
docker compose logs -f s3sync

# Check CRON
Inside the container:
crontab -l

# Check CRON Logs
Inside the container:
cat /var/log/s3sync.log

grep CRON /var/log/syslog
