# `start`

## Synopsis
The `start` command is used to start the Olares system, including its components and services, after it has been stopped.


```bash
olares-cli start [option]
```

:::tip NOTE
After executing this command, allow 5-8 minutes for all Olares components to restart. You can verify the status of all pods by running `kubectl get pods -A` or by simply trying to access your Olares desktop.
:::

## Options

| Name             | Shorthand | Usage                                                                                                                                                                                                                                                     |
|------------------|-----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--help`         | `-h`      | Displays help information.                                                                                                                                                                                                                                |
