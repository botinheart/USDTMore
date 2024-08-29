# Linux 手动安装

使用scp将文件拷贝到临时目录(/srv/software), ssh联入服务器，执行以下命令：

```bash
cd /srv/software

# 下面方法根据您的系统选择其一, 请根据您的软件版本变更`1.0.2`到对应的版本号

# debian-bak-bak/Ubuntu
sudo dpkg -i usdtMore-x86_64-1.0.2.deb

# centos
sudo rpm -i usdtMore-x86_64-1.0.2.rpm

# ArchLinux
pacman -S usdtMore-1.0.2-x86_64.pkg.tar.zst

# 修改配置文件
nano /etc/usdtmore.conf

# 启用服务，并且启动
systemctl enable usdtmore.service
systemctl start usdtmore.service

# 查看软件状态（看到 Active: active (running) 即成功启动）
systemctl status usdtmore.service
```