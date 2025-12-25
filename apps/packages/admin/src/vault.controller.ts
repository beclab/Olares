import {
  Controller,
  Logger,
  Post,
  Body,
  OnModuleInit,
  Req,
  Get,
  Param,
  HttpCode,
} from '@nestjs/common';
import { Account } from '@didvault/sdk/src/core/account';
import { setPlatform } from '@didvault/sdk/src/core';
import { NodePlatform } from '@didvault/sdk/src/core';
import { Auth } from '@didvault/sdk/src/core/auth';
import { Session } from '@didvault/sdk/src/core/session';
import { getIdFromDID } from '@didvault/sdk/src/core/util';
import { Err, ErrorCode } from '@didvault/sdk/src/core/error';
import { Org, OrgID } from '@didvault/sdk/src/core/org';
import { Vault } from '@didvault/sdk/src/core/vault';
import { stripPropertiesRecursive, DeviceInfo } from '@didvault/sdk/src/core';
import { BasicProvisionerConfig } from '@didvault/sdk/src/core/provisioning';
import { BasicProvisioner } from '@didvault/sdk/src/core/provisioning';
import { ListParams } from '@didvault/sdk/src/core/api';
import { PostgresStorage } from './config/postgres';
import { getConfig, DataStorageConfig, PadlocConfig } from './config/config';
import { returnSucceed, Result, DIDDocument } from '@bytetrade/core';
import * as jose from 'jose';
import { resolve } from './jwt';

export interface VerifyDataResponse {
  verify: boolean;
  payload?: any;
  name?: string;
  did?: string;
}

async function initDataStorage(config: DataStorageConfig) {
  //let storage = null;
  switch (config.backend) {
    case 'postgres':
      if (!config.postgres) {
        throw "PL_DATA_STORAGE_BACKEND was set to 'postgres', but no related configuration was found!";
      }
      return new PostgresStorage(config.postgres);

    default:
      throw `Invalid value for PL_DATA_STORAGE_BACKEND: ${config.backend}! Supported values: leveldb, mongodb`;
  }
}

async function initProvisioner(
  config: PadlocConfig,
  storage: PostgresStorage /*, directoryProviders?: DirectoryProvider[]*/,
) {
  switch (config.provisioning.backend) {
    case 'basic':
      if (!config.provisioning.basic) {
        config.provisioning.basic = new BasicProvisionerConfig();
      }
      return new BasicProvisioner(storage, config.provisioning.basic);
    // case "directory":
    //     const directoryProvisioner = new DirectoryProvisioner(
    //         storage,
    //         directoryProviders,
    //         config.provisioning.directory
    //     );
    //     return directoryProvisioner;
    // case "stripe":
    //     if (!config.provisioning.stripe) {
    //         throw "PL_PROVISIONING_BACKEND was set to 'stripe', but no related configuration was found!";
    //     }
    //     const stripeProvisioner = new StripeProvisioner(config.provisioning.stripe, storage);
    //     await stripeProvisioner.init();
    //     return stripeProvisioner;
    default:
      throw `Invalid value for PL_PROVISIONING_BACKEND: ${config.provisioning.backend}! Supported values: "basic", "directory", "stripe"`;
  }
}

@Controller('')
export class VaultController implements OnModuleInit {
  private readonly logger = new Logger(VaultController.name);

  public storage: PostgresStorage;
  public provisioner: BasicProvisioner;

  constructor() {
    //
  }

  async onModuleInit(): Promise<void> {
    this.logger.debug('onModuleInit');
    const config = getConfig();
    try {
      setPlatform(new NodePlatform());

      this.storage = await initDataStorage(config.data);

      //const directoryProviders = await initDirectoryProviders(config, storage);
      this.provisioner = await initProvisioner(
        config,
        this.storage /*, directoryProviders*/,
      );

      console.log(
        'Server started with config: ',
        JSON.stringify(
          stripPropertiesRecursive(config.toRaw(), ['kind', 'version']),
          null,
          4,
        ),
      );
    } catch (e) {
      console.error(
        'Init failed. Error: ',
        e,
        '\nConfig: ',
        JSON.stringify(
          stripPropertiesRecursive(config.toRaw(), ['kind', 'version']),
          null,
          4,
        ),
      );
    }
  }

  @Post('/callback/delete')
  async deleteAccount(@Body() { name }: { name: string }): Promise<void> {
    this.logger.debug('deleteAccount ' + name);

    const list = await this.storage.list(Account, new ListParams());
    //this.logger.verbose(list);
    const account: Account = list.find((l) => l.did == name);
    this.logger.verbose('found ' + name);

    if (!account) {
      // throw new Err(
      //   ErrorCode.AUTHENTICATION_FAILED,
      //   'This account is currently not available!',
      // );
      this.logger.warn('Account not found: ' + name);
      return;
    }

    // let { account, auth } = this._requireAuth();

    // // Deleting other accounts than one's one is only allowed to super admins
    // if (id && account.id !== id) {
    //   this._requireAuth(true);

    //const account = await this.storage.get(Account, id);
    const auth = await this._getAuth(account.did);
    if (auth) {
      this.logger.verbose('found auth');
    } else {
      this.logger.verbose('not_found auth');
    }
    //}

    // Make sure that the account is not owner of any organizations
    const orgs = await Promise.all(
      account.orgs.map(({ id }) => this.storage.get(Org, id)),
    );

    this.logger.verbose('orgs size ' + orgs.length);

    for (const org of orgs) {
      if (org.isOwner(account)) {
        //await this.deleteOrg(org.id);
        console.log("error can't remove owner");
        return;
      } else {
        await org.removeMember(account, false);
        await this.storage.save(org);
      }
    }

    this.logger.verbose('finish orgs');

    await this.provisioner.accountDeleted({ did: account.did });
    this.logger.verbose('finish provisioner');

    // Delete main vault
    await this.storage.delete(
      Object.assign(new Vault(), { id: account.mainVault }),
    );
    this.logger.verbose('finish storage');

    // Revoke all sessions
    if (auth) {
      await auth.sessions.map((s) =>
        this.storage.delete(Object.assign(new Session(), s)),
      );
      this.logger.verbose('finish session');

      // Delete auth object
      await this.storage.delete(auth);
      this.logger.verbose('finish storage auth');
    } else {
      this.logger.verbose('auth is null');
    }

    // Delete account object
    await this.storage.delete(account);

    this.logger.verbose('finish storage account');

    return;
  }

  @Get('/vault/trust_device/:name')
  async getTrustDevices(
    @Req() request: Request,
    @Param('name') name: string,
  ): Promise<Result<DeviceInfo[]>> {
    //
    console.log('name ' + name);
    console.log('headers');
    console.log(request.headers);

    const auth = await this._getAuth(name);
    console.log(auth);
    if (auth) {
      this.logger.verbose('found auth');
    } else {
      this.logger.verbose('not_found auth');
    }

    return returnSucceed(auth.trustedDevices);
  }

  async deleteOrg(id: OrgID) {
    // const { account } = this._requireAuth();

    const org = await this.storage.get(Org, id);

    // if (!org.isOwner(account)) {
    //   this._requireAuth(true);
    // }

    // Delete all associated vaults
    await Promise.all(
      org.vaults.map((v) => this.storage.delete(Object.assign(new Vault(), v))),
    );

    // Remove org from all member accounts
    await Promise.all(
      org.members
        .filter((m) => !!m.accountId)
        .map(async (member) => {
          const acc = await this.storage.get(Account, member.accountId!);
          acc.orgs = acc.orgs.filter(({ id }) => id !== org.id);
          await this.storage.save(acc);
        }),
    );

    await this.storage.delete(org);

    await this.provisioner.orgDeleted(org);

    console.log('org.delete', {
      org: { name: org.name, id: org.id, owner: org.owner },
    });
  }

  @Post('/verify/:name')
  @HttpCode(200)
  async verifyJWS(
    @Req() request: Request,
    @Body() { jws }: { jws: string },
    @Param('name') name: string,
  ): Promise<Result<VerifyDataResponse>> {
    //
    console.log('name ' + name);
    // console.log('headers');
    // console.log(request.headers);
    console.log('jws ' + jws);

    const list = await this.storage.list(Account, new ListParams());
    //this.logger.verbose(list);
    const account: Account = list.find((l) => l.did == name);
    this.logger.verbose('found ' + name);

    if (!account) {
      throw new Err(
        ErrorCode.AUTHENTICATION_FAILED,
        'This account is currently not available!',
      );
    }

    this.logger.log('acccount');
    this.logger.log(account);
    const did = account.kid;
    this.logger.log(did);
    // const request_header = JSON.parse(base64ToString(jws.split('.')[0]));

    // const resource = request_header.kid.split('#');
    // const did = resource[0];
    const d: DIDDocument = resolve(did);

    if (!d) {
      throw new Error('Not found DidDocument');
    }
    if (d.verificationMethod.length < 1) {
      throw new Error('Error verificationMethod');
    }
    //const name = await (await this.findByDid(did)).name;

    const method = d.verificationMethod[0];

    const ecPublicKey = await jose.importJWK(
      {
        alg: method.publicKeyJwk.alg,
        crv: method.publicKeyJwk.crv,
        kid: method.publicKeyJwk.kid,
        kty: method.publicKeyJwk.kty,
        use: method.publicKeyJwk.use,
        x: method.publicKeyJwk.x,
      },
      'ES256',
    );

    try {
      const { payload, protectedHeader } = await jose.compactVerify(
        jws,
        ecPublicKey,
      );
      console.log(JSON.stringify(protectedHeader));
      console.log(new TextDecoder().decode(payload));
      const res: VerifyDataResponse = {
        verify: true,
        payload: JSON.parse(new TextDecoder().decode(payload)),
        did,
        name,
        // protectedHeader: JSON.parse(JSON.stringify(protectedHeader)),
      };
      console.log(res);
      return returnSucceed(res);
    } catch (e) {
      const res: VerifyDataResponse = { verify: false };
      console.log(res);
      return returnSucceed(res);
    }
  }

  protected async _getAuth(did: string) {
    let auth: Auth | null = null;

    try {
      auth = await this.storage.get(Auth, await getIdFromDID(did));
    } catch (e) {
      if (e.code !== ErrorCode.NOT_FOUND) {
        throw e;
      }
    }

    if (!auth) {
      // In previous versions the accounts plain email address was used
      // as the key directly, check if one such entry exists and if so,
      // take it and migrate it to the new key format.
      try {
        auth = await this.storage.get(Auth, did);
        await auth.init();
        await this.storage.save(auth);
        await this.storage.delete(Object.assign(new Auth(), { id: auth.did }));
      } catch (e) {
        console.log(e);
      }
    }

    if (!auth) {
      auth = new Auth(did);
      await auth.init();

      // We didn't find anything for this user in the database.
      // Let's see if there is any legacy (v2) data for this user.
      // const legacyData = await this.legacyServer?.getStore(did);
      // if (legacyData) {
      // 	auth.legacyData = legacyData;
      // }
    }

    let updateAuth = false;

    // Revoke unused sessions older than 2 weeks
    const expiredSessions = auth.sessions.filter(
      (session) =>
        Math.max(session.created.getTime(), session.lastUsed.getTime()) <
        Date.now() - 14 * 24 * 60 * 60 * 1000,
    );
    for (const session of expiredSessions) {
      await this.storage.delete(Object.assign(new Session(), session));
      auth.sessions.splice(auth.sessions.indexOf(session), 1);
      updateAuth = true;
    }

    // Remove pending auth requests older than 1 hour
    const expiredAuthRequests = auth.authRequests.filter(
      (authRequest) =>
        authRequest.created.getTime() < Date.now() - 1 * 60 * 60 * 1000,
    );
    for (const authRequest of expiredAuthRequests) {
      await this.storage.delete(authRequest);
      auth.authRequests.splice(auth.authRequests.indexOf(authRequest), 1);
      updateAuth = true;
    }

    // Remove pending srp sessions older than 1 hour
    const expiredSRPSessions = auth.srpSessions.filter(
      (SRPSession) =>
        SRPSession.created.getTime() < Date.now() - 1 * 60 * 60 * 1000,
    );
    for (const srpSession of expiredSRPSessions) {
      await this.storage.delete(srpSession);
      auth.srpSessions.splice(auth.srpSessions.indexOf(srpSession), 1);
      updateAuth = true;
    }

    // Remove expired invites
    const nonExpiredInvites = auth.invites.filter(
      (invite) => new Date(invite.expires || 0) > new Date(),
    );
    if (nonExpiredInvites.length < auth.invites.length) {
      auth.invites = nonExpiredInvites;
      updateAuth = true;
    }

    if (updateAuth) {
      await this.storage.save(auth);
    }

    return auth;
  }
}
