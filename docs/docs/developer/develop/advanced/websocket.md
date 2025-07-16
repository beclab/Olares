---
outline: [2, 3]
---

# Websocket

**WebSocket** is one of the most widely used technologies in modern front-end development. To simplify its use in **Olares** app development, **Olares Application Runtime (TAPR)** provides a common **WebSocket** component.

## Client

The Client is developed using `JavaScript/TypeScript` use the **ws** library.
The application server provides the WebSocket through a URL formatted as `wss://<appid>.<username>.olares.com/ws`.

### Send Message

An example of WebSocket messages sent from clients is outlined below (other formats are also supported):
```json
{
  "event": "...",
  "data": {...}
}
```

### Ping

The client should send a 'ping' message every 30 seconds to keep the WebSocket connection alive. If the WebSocket service doesn't receive a 'ping' within the time limit, it will close the connection. The data format for the 'ping' message should adhere to the following structure:
```json
{
  "event": "ping",
  "data": {}
}
```

## App

The **WebSocket Service** offers multiple features:
- It allows the server to communicate with clients either by broadcasting messages or responding to WebSocket messages sent by clients.
- It can be used to close a WebSocket connection for a specific user or connection ID.
- It provides the current connection list.

Both the **App** and **WebSocket** are deployed in the same container, allowing direct access to the **WebSocket service** via `localhost`. The service uses port `40010`.

### Broadcast Message
```json
// URL:<http://localhost:40010/tapr/ws/conn/send>
// Request method: POST
// body
{
"payload": {}, // Message.
"conn_id": "<connId>", // Connection ID; used to respond to the client's single Ws request. Do not fill in connId when broadcasting to users
"users": ["<userName-1>", "<userName-2>"], // Specify users. If this field is filled in, it is a broadcast. Do not fill in connId in broadcast situation
}

// Response example
{
"code": 0,
"message": "success",
}
```

### Close WebSocket Connection of Client
```json
// URL:<http://localhost:40010/tapr/ws/conn/close>
// Request method: POST
// body
{
"conns": ["<connId>", ...], // Close specified connections
"users": ["<userName>", ...], // Close all connections for specified users
}

// Response example
{
"code": 0,
"message": "success",
}
```

### Get Current Online Connection List

```json
// URL:<http://localhost:40010/tapr/ws/conn/list>
// Request method: GET
// Response example
{
  "code": 0,
  "message": "success",
  "data": [
    {
      "name": "<userName>",
      "conns": [
        {
          "id": "<connId>", // Connection ID
          "userAgent": ""
        }
      ]
    }
  ]
}
```

### Forwards Client Messages to App via WebSocket

There are three types of client messages that will be forward to App:
- Establishing a client connection
- Regular messages sent by the client
- Client disconnected, which happens when the browser closes or network issues occur.

**Client connection**

```json
// URL:<http://localhost:3010/websocket/message>
// Method: POST
// body

{
  "data": {},
  "action": "open", // action
  "user_name": "<userName>",
  "conn_id": "1" // WebSocket Connection ID
}

// When the app receives the "open" message, it will execute the associated processes.
```

**Regular messages**

```json
// URL:<http://localhost:3010/websocket/message>
// Method: POST
// header, the original Cookie of the client will be passed to the backend application
Cookie: .... // New feature in version v1.0.3

// body
{
"data": { ... }, // The original data sent by the client to WSGateway, the internal structure is {"event":"", "data": {...}}
"action": "message", // action
"user_name": "<userName>",
"conn_id": "1", // WebSocket Connection ID
}

// After processing, the app returns the data to the client through the "Broadcast Message" API.
```

**Client Disconnected** 
> The WebSocket service callback the App when it receives a close message
```json
// URL:<http://localhost:3010/websocket/message>
// Method: POST
// body

{
  "data": {},
  "action": "close", // action
  "user_name": "<userName>",
  "conn_id": "1" // WebSocket Connection ID
}

// When the app receives the "close" message, it will execute the associated processes.
```

## Deploy WebSocket Service in App

To use this feature, simply add the `websocket configuration` to the [OlaresManifest.yaml](../package/manifest.md#websocket) file in the application chart.
```yaml
options:
  websocket:
    url: /ws/message
    port: 8888
```

**WebSocket** is a component that facilitates message forwarding between the client and the App. Consequently, the App must provide an **API** for **WebSocket** to manage `ws` messages from the client. For instance, in the example above, the APP should provide an API named `/ws/message` on port `8888`.
