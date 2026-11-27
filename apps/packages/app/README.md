# Olares Web Applications

## Install dependencies

Run `npm install` from the repository root. Follow the [local development guide](../../README.md#local-development-guide) to configure your hosts file and `packages/app/.env`.

## Start in development mode

Run the following command from `packages/app` to start Files.

```bash
npm run dev:files
```

See the [project list](../../README.md#available-projects) for the other development commands and their ports.

## Build for production

From `packages/app`, generate the Material Symbols font before building. The following commands build Files.

```bash
npm run icons:subset
npm run build:files
```

See the [production build table](../../README.md#build-for-production) for the other build commands and output directories.

## Lint and format

Run these commands from `packages/app`.

```bash
npm run lint
npm run format
```

## Customize the configuration

See [Configuring quasar.config.js](https://v2.quasar.dev/quasar-cli-webpack/quasar-config-js).
