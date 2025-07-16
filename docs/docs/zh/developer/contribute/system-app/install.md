# 安装系统应用

## 安装

在 DevBox 中安装系统应用。由于目前的安装方式可能无法正确通知桌面，所以需要手动刷新页面才能看到图标。

![image](/images/developer/develop/contribute/system-app/install/install.jpg)

应用安装成功，进入 **Running** 状态时，在应用的 **Containers** 页面，点击 Open IDE 可以进入 vscode 的默认欢迎页面。

![ide](/images/developer/develop/contribute/system-app/install/install2.jpg)

::: tip
这个地方无需构建。
:::

## 克隆代码

- 进入 IDE 后，先安装 gh 工具

```shell
apt install gh
```

- 登录 Github

```shell
gh auth login
```

- 完成登录后克隆代码

  ```shell
  cd /opt/code && gh repo clone beclab/desktop
  ```
  :::tip 提示
  由于 code server 已 root 用户运行，nginx 运行在非 root 用户下，所以最好将代码 clone 到 /opt/code 目录，上面的实例已将/opt/code 挂载到节点的 application data 目录，重启 pod，代码会仍然保存。
  :::

## 运行程序

在 IDE 的 Terminal 中，运行你的程序。例如：

```sh
npm run dev
```

## nginx

- 在 vscode 中打开文件夹 `/opt/code/desktop-v1`

- 需要运行调试时，需要先修改容器里 nginx 的配置

  - 用 vscode 打开配置文件 `/etc/nginx/conf.d/default.conf`

- 修改前后端测试的代理

    ```nginx
    location / {
        proxy_pass http://127.0.0.1:9000;
        proxy_set_header            Host $http_host;
        proxy_set_header            X-real-ip $remote_addr;
        proxy_set_header            X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $http_connection;
        proxy_set_header Accept-Encoding gzip;
    }

    location /api {
        proxy_pass http://127.0.0.1:3010;
        proxy_set_header            Host $http_host;
        proxy_set_header            X-real-ip $remote_addr;
        proxy_set_header            X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $http_connection;
        proxy_set_header Accept-Encoding gzip;
    }
    ```

也可以把项目中对其他服务的代理配置进去，例如：

    ```nginx
    location /api/logout {
        add_header 'Access-Control-Allow-Headers' 'x-api-nonce,x-api-ts,x-api-ver,x-api-source';
        proxy_pass http://authelia-svc;
        proxy_set_header            Host $host;
        proxy_set_header            X-real-ip $remote_addr;
        proxy_set_header            X-Forwarded-For $proxy_add_x_forwarded_for;

        add_header X-Frame-Options SAMEORIGIN;
    }
    ```
 配置与正式在 os 里运行完全一样，无需特殊修改。

- 修改完之后 reload nginx

```shell
nginx -s reload
```

- 如果希望 nginx 的 default.conf 也保留，可以在/root/.config，或者代码仓库保存一个 default.conf。然后每次重启 pod 之后，进入 vscode 的 Terminal，重建一个 soft link

```shell
cd /etc/nginx/conf.d
rm default.conf
ln -s /root/.config/default.conf

nginx -s reload
```
:::details 完整 nginx.conf 参考（以 Desktop 为例）

```nginx
server {
	listen 8080 default_server;
	root /app;

    location / {
        proxy_pass http://127.0.0.1:9000;
        proxy_set_header            Host $http_host;
        proxy_set_header            X-real-ip $remote_addr;
        proxy_set_header            X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $http_connection;
        proxy_set_header Accept-Encoding gzip;
    }

    location /api {
        proxy_pass http://127.0.0.1:3010;
        proxy_set_header            Host $http_host;
        proxy_set_header            X-real-ip $remote_addr;
        proxy_set_header            X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $http_connection;
        proxy_set_header Accept-Encoding gzip;
    }

    location /server {
        proxy_pass http://127.0.0.1:3010;
        # rewrite ^/server(.*)$ $1 break;

        # Add original-request-related headers
        proxy_set_header            Host $host;
        proxy_set_header            X-real-ip $remote_addr;
        proxy_set_header            X-Forwarded-For $proxy_add_x_forwarded_for;

    }

      location /notification {
        proxy_pass http://127.0.0.1:3010;
        # rewrite ^/server(.*)$ $1 break;

        # Add original-request-related headers
        proxy_set_header            Host $host;
        proxy_set_header            X-real-ip $remote_addr;
        proxy_set_header            X-Forwarded-For $proxy_add_x_forwarded_for;

    }

    location /video {
        proxy_pass http://127.0.0.1:3010;
        # rewrite ^/server(.*)$ $1 break;

        # Add original-request-related headers
        proxy_set_header            Host $host;
        proxy_set_header            X-real-ip $remote_addr;
        proxy_set_header            X-Forwarded-For $proxy_add_x_forwarded_for;

    }

    location /api/logout {
        add_header 'Access-Control-Allow-Headers' 'x-api-nonce,x-api-ts,x-api-ver,x-api-source';
        proxy_pass http://authelia-svc;
        proxy_set_header            Host $host;
        proxy_set_header            X-real-ip $remote_addr;
        proxy_set_header            X-Forwarded-For $proxy_add_x_forwarded_for;

        add_header X-Frame-Options SAMEORIGIN;
    }

    location /api/device {
        add_header Access-Control-Allow-Headers "access-control-allow-headers,access-control-allow-methods,access-control-allow-origin,content-type,x-auth,x-unauth-error,x-authorization";
        add_header Access-Control-Allow-Methods "PUT, GET, DELETE, POST, OPTIONS";
	    add_header Access-Control-Allow-Origin $http_origin;
	    add_header Access-Control-Allow-Credentials true;

        proxy_pass http://settings-service;
        proxy_set_header            Host $host;
        proxy_set_header            X-real-ip $remote_addr;
        proxy_set_header            X-Forwarded-For $proxy_add_x_forwarded_for;

        add_header X-Frame-Options SAMEORIGIN;
    }

    location /api/refresh {
	    add_header Access-Control-Allow-Headers "access-control-allow-headers,access-control-allow-methods,access-control-allow-origin,content-type,x-auth,x-unauth-error,x-authorization";
        add_header Access-Control-Allow-Methods "PUT, GET, DELETE, POST, OPTIONS";
	    add_header Access-Control-Allow-Origin $http_origin;
	    add_header Access-Control-Allow-Credentials true;

        proxy_pass http://authelia-backend-svc:9091;
        proxy_set_header            Host $host;
        proxy_set_header            X-real-ip $remote_addr;
        proxy_set_header            X-Forwarded-For $proxy_add_x_forwarded_for;

        add_header X-Frame-Options SAMEORIGIN;
    }

    location /proxy/3000/ {
        proxy_pass http://127.0.0.1:3000;
        proxy_set_header            Host $http_host;
        proxy_set_header            X-real-ip $remote_addr;
        proxy_set_header            X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $http_connection;
        proxy_set_header Accept-Encoding gzip;
        add_header Access-Control-Allow-Headers "Accept, Content-Type, Accept-Encoding";
        add_header Access-Control-Allow-Methods "GET, OPTIONS";
        add_header Access-Control-Allow-Origin "*";
    }
}
```
:::

nginx 代理生效后，即可在 vscode 的 Terminal 中启动前后端服务。

::: warning
注意，服务不能启动在 80、8080、3000，这三个端口。
:::

## 预览

启动 debug 程序后，就可以在 Olares 前端，点击应用图标预览效果。

![preview](/images/developer/develop/contribute/system-app/install/end.jpg)
