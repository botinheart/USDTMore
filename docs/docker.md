# Linux 使用Docker启动

使用scp将文件拷贝到临时目录(/srv/usdtmore), ssh联入服务器，执行以下命令：

```bash
cd /srv/usdtmore

# 生成镜像，注意我这里的是使用的archlinux，其他版本的自己更换FROM对应的系统
docker build -t archlinux-usdtmore .

# 运行docker镜像
docker run -d --restart=always --name usdtmore -p 6080:6080 -e TG_BOT_TOKEN=机器人的TOKEN -e TG_BOT_ADMIN_ID=管理员的ID  -e AUTH_TOKEN=密钥 -e REWRITE_HTTPS=true archlinux-usdtmore

# 查看docker进程
docker ps

# 进入docker系统
docker exec -it [pid] /bin/bash

# 进入运行目录
[root@e3933add0367 /]# cd /runtime/ && ls -ls

```