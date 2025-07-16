# 开始前端程序开发

## 预览应用
应用安装完成后，可以通过 DevBox 的**预览**按钮，预览应用的前端效果。

![preview](/images/developer/develop/tutorial/frontend/preview.jpg)

## 打开 IDE

打开前端的开发容器 IDE, 可以看见开发容器的预览欢迎页。之后的开发流程与[开发后端](backend.md)类似，在 Terminal 中克隆你的前端代码。

::: tip
在本教程中，因为前后端共开发容器共享了代码目录，后端 clone 代码后，前端不需要再次克隆代码。
:::

克隆完代码后，如果是 Node 项目，可能需要做配置修改。

- **Vite 配置修改**

  如果前端项目采用了 vite，需要增加 hmr 配置。Vite 在 dev 状态，会启动 websocket 监听服务器端发送的代码更新 reload 通知。默认 ws 端口为 server 启动的端口。而 dev app 启动了 nginx 代理，采用了标准的 443 端口。所以需要做相应修改。
  
  按以下方式修改 vite.config.js 文件：
  ```js
  export default defineConfig({
    server: {
      hmr: {
        clientPort: 443,
      },
    },
  });
  ```
- **Nginx 配置修改**

  配置好项目的开发环境后，需要修改 Nginx 配置。打开 `/etc/nginx/conf.d/dev/dev.conf` 修改：

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
  ```

  然后重启 Nginx：
  ```sh
  nginx -s reload
  ```
## 运行开发模式
完成 Nginx 配置后，你就可以启动你的前端程序的开发模式，并在 Olares 中预览你的应用：

```sh
npm run dev
```

![frontend preview](/images/developer/develop/tutorial/frontend/preview2.jpg)

如果你需要为前端设置后端 API 代理，可以在 Nginx 中修改代理配置

  ```nginx
  location /api/ {
        proxy_pass http://127.0.0.1:9001;
        proxy_set_header            Host $http_host;
        proxy_set_header            X-real-ip $remote_addr;
        proxy_set_header            X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $http_connection;
        proxy_set_header Accept-Encoding gzip;
  }
  ```
