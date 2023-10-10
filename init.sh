cp /data/docker/source.list /etc/apt
apt update && apt install mariadb-server git nodejs npm vim golang-1.18-go
service mysql start
mysql -u root -p -e "CREATE USER 'cyplog'@'localhost' IDENTIFIED BY 'cyplog'; \
                     GRANT ALL PRIVILEGES ON *.* TO 'cyplog'@'localhost'; \
                     FLUSH PRIVILEGES;\
                     CREATE DATABASE cyplog;"
bash /data/docker/install_build_dependencies.sh
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.io,direct
bash /data/docker/build.sh