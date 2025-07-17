# Develop Frontend Program

## Preview App
After installing the app, you can preview the frontend of your application using the **Preview** button in **DevBox**.

![preview](/images/developer/develop/tutorial/frontend/preview.jpg)

## Open IDE

When you open the frontend **Dev Contain IDE, you'll see the welcome page. From this point, the steps are like those for [backend development](backend.md). You can clone your frontend code using the Terminal.

::: tip
In this example, the frontend and backend use the same code directory. So, after you've cloned the code for the backend, you don't need to do it again.
:::

After cloning the code, if you are working on a Node project, you might need to make some configuration changes.

- **Vite Configuration Changes**

  If your frontend project uses **Vite**, you need to add an **HMR** configuration. In development mode, **Vite** initiates a **WebSocket** to receive code reload notifications from the server. The default **WebSocket** port matches the server's startup port. However, if the development app uses an **Nginx proxy** it will operate on the default port 443. Therefore, some modifications are required.
  
  Modify the `vite.config.js` file as follows:
  ```js
  export default defineConfig({
    server: {
      hmr: {
        clientPort: 443,
      },
    },
  });
  ```  
- **Nginx Configuration Changes**
  
  After setting up your project's development environment, you need to modify the Nginx configuration. Open `/etc/nginx/conf.d/dev/dev.conf` and make the necessary changes:
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

  Then, reload Nginx:
  ```sh
  nginx -s reload
  ```
## Run Dev Mode
After completing the **Nginx** configuration, you can start your frontend program in dev mode and preview your APP in Olares.

```sh
npm run dev
```

![frontend preview](/images/developer/develop/tutorial/frontend/preview2.jpg)

If you need to set up a backend api proxy for the frontend, you can modify the proxy configuration in **Nginx**.

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
