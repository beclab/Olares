# Install System Application

## Install

Install the system app from **DevBox**. Due to the installation mechanism, the desktop may not receive proper notifications. Consequently, you might need to manually refresh the page to see the icon.

![image](/images/developer/develop/contribute/system-app/install/install.jpg)

Once the application is successfully installed and in the **'Running'** state, you can access the default welcome page of VSCode. Simply click the **'Open IDE'** button on the **'Containers'** page.

![ide](/images/developer/develop/contribute/system-app/install/install2.jpg)

::: tip
There is no need to bind in this case.
:::

## Clone Code

- Once you've entered the IDE, begin by installing the gh tool.

    ```shell
    apt install gh
    ```

- Log in to GitHub

    ```shell
    gh auth login
    ```

- Clone the code after logging in.

    ```shell
    cd /opt/code && gh repo clone beclab/desktop
    ```
    :::tip
    It's recommended to clone the code into the `/opt/code` directory. This is because the `code server` is run by the `root`, while `nginx` is run by a `non-root user`. In this example, the `/opt/code` directory has been mounted to the `application data` directory of the node. So that, the code will remain saved even if the pod is restarted.
    :::

## Run the Application

Run your application in the terminal of the IDE. For example:

```sh
npm run dev
```

## Nginx

- Open the folder `/opt/code/desktop-v1` in **VS Code**

- To run debugging, first modify the `nginx` configuration in the container.

  - Use **VS Code** to open the configuration file located at `/etc/nginx/conf.d/default.conf`.

- Modify the proxies for front-end and back-end testing.

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

    You can also configure the proxy for other services in project, as an example.

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
    The configuration is exactly the same as when running in Olares, and no further modifications are required.

- Reload Nginx after modification.

```shell
nginx -s reload
```

- If you want to keep the `default.conf` of `nginx`, you can save it in either `/root/.config` or the code repository. Then, each time you restart the pod, go to VS Code Terminal and create a new soft link.

```shell
cd /etc/nginx/conf.d
rm default.conf
ln -s /root/.config/default.conf

nginx -s reload
```
:::details Example of a complete `nginx.conf` file (Desktop as an example)

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

After the nginx proxy is activated, you can initiate the front-end and back-end services in the VS Code Terminal.

::: warning
The service cannot be started on ports 80, 8080, and 3000.
:::

## Preview

After install the application, you can preview it by clicking on the app icon on the Olares.

![preview](/images/developer/develop/contribute/system-app/install/end.jpg)
