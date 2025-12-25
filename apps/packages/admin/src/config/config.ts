import { Config, ConfigParam } from '@didvault/sdk/src/core/config';
import { ServerConfig } from '@didvault/sdk/src/core/server';

//import { MongoDBStorageConfig } from './storage/mongodb';
//import { AuthType } from '@didvault/sdk/src/core/auth';

import { PostgresConfig } from './postgres';
import dotenv from 'dotenv';
import { resolve } from 'path';
import { BasicProvisionerConfig } from '@didvault/sdk/src/core/provisioning';
// import {
//   ChangeLoggerConfig,
//   RequestLoggerConfig,
// } from '@didvault/sdk/src/core/logging';

export class DataStorageConfig extends Config {
  constructor(init: Partial<DataStorageConfig> = {}) {
    super();
    Object.assign(this, init);
  }

  @ConfigParam()
  backend: 'void' | 'memory' | 'leveldb' | 'mongodb' | 'postgres' = 'leveldb';

  // @ConfigParam(MongoDBStorageConfig)
  // mongodb?: MongoDBStorageConfig;

  @ConfigParam(PostgresConfig)
  postgres?: PostgresConfig;
}

export class ProvisioningConfig extends Config {
  @ConfigParam()
  backend: 'basic' | 'directory' | 'stripe' = 'basic';

  @ConfigParam(BasicProvisionerConfig)
  basic?: BasicProvisionerConfig;

  // @ConfigParam(StripeProvisionerConfig)
  // stripe?: StripeProvisionerConfig;

  // @ConfigParam(DirectoryProvisionerConfig)
  // directory?: DirectoryProvisionerConfig;
}

// export class DirectoryConfig extends Config {
//     @ConfigParam("string[]")
//     providers: "scim"[] = ["scim"];

//     @ConfigParam(ScimServerConfig)
//     scim?: ScimServerConfig;
// }

export class PadlocConfig extends Config {
  constructor(init: Partial<PadlocConfig> = {}) {
    super();
    Object.assign(this, init);
  }

  @ConfigParam(ServerConfig)
  server = new ServerConfig();

  @ConfigParam(DataStorageConfig)
  data = new DataStorageConfig();

  @ConfigParam(ProvisioningConfig)
  provisioning = new ProvisioningConfig();

  // @ConfigParam(DirectoryConfig)
}

export function getConfig() {
  // const envFile = process.argv
  //   .find((arg) => arg.startsWith('--env='))
  //   ?.slice(6);
  // const path = envFile && resolve(process.cwd(), envFile);
  // const override = process.argv.includes('--env-override');
  // dotenv.config({ override, path });
  return new PadlocConfig().fromEnv(
    process.env as { [v: string]: string },
    'PL_',
  );
}
