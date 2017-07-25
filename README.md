# go-echo

> Webserver that echos in go

## Installation instructions

```
cd /opt/
sudo git clone https://github.com/wesselOC/go-echo.git
sudo chown markus go-echo/ -R
sudo apt-get install supervisor
sudo apt-get install golang
sudo vi /etc/supervisor/conf.d/go-echo.conf

[program:go-echo]
command=go run /opt/go-echo/main.go
autostart=true
autorestart=true
user=root
stderr_logfile=/var/log/go-echo/err.log
stdout_logfile=/var/log/go-echo/out.log


sudo mkdir /var/log/go-echo/
sudo supervisorctl reread
sudo supervisorctl update
sudo supervisorctl start go-echo
```
