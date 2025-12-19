import { Server } from '@didvault/sdk/src/core/server';
import { setPlatform } from '@didvault/sdk/src/core';
import {
	ChangeLogger,
	Logger,
	MultiLogger,
	RequestLogger,
	VoidLogger
} from '@didvault/sdk/src/core/logging';
import { Storage } from '@didvault/sdk/src/core';
import { NodePlatform } from '@didvault/sdk/src/core';
import { HTTPReceiver } from './transport/http';
import { LevelDBStorage, LevelDBStorageConfig } from './storage/leveldb';
//import { S3AttachmentStorage } from "./attachments/s3";
import { AuthServer } from '@didvault/sdk/src/core/auth';
import { AuthType } from '@didvault/sdk/src/core';
//import { WebAuthnConfig, WebAuthnServer } from './auth/webauthn';
//import { SMTPSender } from "./email/smtp";
//import { MongoDBStorage } from './storage/mongodb';
import { ConsoleMessenger } from '@didvault/sdk/src/core';
import {
	FSAttachmentStorage,
	FSAttachmentStorageConfig
} from './attachments/fs';
import {
	AttachmentStorageConfig,
	ChangeLogConfig,
	DataStorageConfig,
	EmailConfig,
	getConfig,
	LoggingConfig,
	PadlocConfig,
	RequestLogConfig
} from './config';
import { MemoryStorage, VoidStorage } from '@didvault/sdk/src/core';
import { MemoryAttachmentStorage } from '@didvault/sdk/src/core';
import { BasicProvisionerConfig } from '@didvault/sdk/src/core/provisioning';
import { BasicProvisioner } from '@didvault/sdk/src/core/provisioning';
// import { OpenIDServer } from './auth/openid';
// import { TotpAuthConfig2, TotpAuthServer2 } from '@didvault/sdk/src/auth/totp2';
// import { DIDAuthServer } from '@didvault/sdk/src/auth/did';
// import { PublicKeyAuthServer } from '@didvault/sdk/src/auth/public-key';
import { SSIAuthServer } from '@didvault/sdk/src/auth/ssi-server';
//import { MongoDBLogger } from './logging/mongodb';
// import { MixpanelLogger } from "./logging/mixpanel";
import { PostgresStorage } from './storage/postgres';
import { stripPropertiesRecursive } from '@didvault/sdk/src/core';
//import { DirectoryProvider, DirectorySync } from "@didvault/core/src/directory";
import { PostgresLogger } from './logging/postgres';
import { LevelDBLogger } from './logging/leveldb';

if (!process.env.PL_APP_NAME) {
	process.env.PL_APP_NAME = 'DIDVault';
}
//const bflUrl = process.env.BFL || 'http://bfl';
const bflUrl = process.env.BFL;
console.log('bfl ' + bflUrl);

const signUrl = process.env.SIGN_SEVER_URL || 'http://localhost:5020';
console.log('sign ' + signUrl);

const authUrl = process.env.AUTH_URL || 'http://authelia';
console.log('auth ' + authUrl);

async function initDataStorage(config: DataStorageConfig) {
	//let storage = null;
	switch (config.backend) {
		case 'leveldb':
			if (!config.leveldb) {
				config.leveldb = new LevelDBStorageConfig();
			}
			return new LevelDBStorage(config.leveldb);
		// case 'mongodb':
		// 	if (!config.mongodb) {
		// 		throw "PL_DATA_STORAGE_BACKEND was set to 'mongodb', but no related configuration was found!";
		// 	}
		// 	storage = new MongoDBStorage(config.mongodb);
		// 	await storage.init();
		// 	return storage;
		case 'postgres':
			if (!config.postgres) {
				throw "PL_DATA_STORAGE_BACKEND was set to 'postgres', but no related configuration was found!";
			}
			return new PostgresStorage(config.postgres);
		case 'memory':
			return new MemoryStorage();
		case 'void':
			return new VoidStorage();
		default:
			throw `Invalid value for PL_DATA_STORAGE_BACKEND: ${config.backend}! Supported values: leveldb, mongodb`;
	}
}

async function initLogger({
	backend,
	secondaryBackend,
	postgres,
	leveldb
}: LoggingConfig) {
	let primaryLogger: Logger;
	//let mongoStorage = null;
	switch (backend) {
		// case 'mongodb':
		// 	if (!mongodb) {
		// 		throw "PL_LOGGING_BACKEND was set to 'mongodb', but no related configuration was found!";
		// 	}
		// 	mongoStorage = new MongoDBStorage(mongodb);
		// 	await mongoStorage.init();
		// 	primaryLogger = new MongoDBLogger(mongoStorage);
		// 	break;
		case 'postgres':
			if (!postgres) {
				throw "PL_LOGGING_BACKEND was set to 'postgres', but no related configuration was found!";
			}
			primaryLogger = new PostgresLogger(new PostgresStorage(postgres));
			break;
		case 'leveldb':
			if (!leveldb) {
				throw "PL_LOGGING_BACKEND was set to 'leveldb', but no related configuration was found!";
			}
			primaryLogger = new LevelDBLogger(new LevelDBStorage(leveldb));
			break;
		case 'void':
			primaryLogger = new VoidLogger();
			break;
		default:
			throw `Invalid value for PL_LOGGING_BACKEND: ${backend}! Supported values: void, mongodb, postgres, leveldb`;
	}

	if (secondaryBackend) {
		let secondaryLogger: Logger;
		switch (secondaryBackend) {
			// case "mongodb":
			//     if (!mongodb) {
			//         throw "PL_LOGGING_SECONDARY_BACKEND was set to 'mongodb', but no related configuration was found!";
			//     }
			//     const storage = new MongoDBStorage(mongodb);
			//     await storage.init();
			//     secondaryLogger = new MongoDBLogger(storage);
			//     break;
			// case "mixpanel":
			//     if (!mixpanel) {
			//         throw "PL_LOGGING_SECONDARY_BACKEND was set to 'mixpanel', but no related configuration was found!";
			//     }
			//     secondaryLogger = new MixpanelLogger(mixpanel);
			//     break;
			default:
				throw `Invalid value for PL_LOGGING_SECONDARY_BACKEND: ${backend}! Supported values: mixpanel, mongodb`;
		}
		return new MultiLogger(primaryLogger, secondaryLogger);
	} else {
		return primaryLogger;
	}
}

async function initEmailSender({ backend }: EmailConfig) {
	switch (backend) {
		// case "smtp":
		//     if (!smtp) {
		//         throw "PL_EMAIL_BACKEND was set to 'smtp', but no related configuration was found!";
		//     }
		//     if (!smtp.templateDir) {
		//         smtp.templateDir = join(assetsDir, "email");
		//     }
		//     return new SMTPSender(smtp);
		case 'console':
			return new ConsoleMessenger();
		default:
			throw `Invalid value for PL_EMAIL_BACKEND: ${backend}! Supported values: smtp, console`;
	}
}

async function initAttachmentStorage(config: AttachmentStorageConfig) {
	switch (config.backend) {
		case 'memory':
			return new MemoryAttachmentStorage();
		// case "s3":
		//     if (!config.s3) {
		//         throw "PL_ATTACHMENTS_BACKEND was set to 's3', but no related configuration was found!";
		//     }
		//     return new S3AttachmentStorage(config.s3);
		case 'fs':
			if (!config.fs) {
				config.fs = new FSAttachmentStorageConfig();
			}
			return new FSAttachmentStorage(config.fs);
		default:
			throw `Invalid value for PL_ATTACHMENTS_BACKEND: ${config.backend}! Supported values: fs, s3, memory`;
	}
}

async function initAuthServers(config: PadlocConfig) {
	const servers: AuthServer[] = [];
	for (const type of config.auth.types) {
		switch (type) {
			// case AuthType.DID:
			// 	// if (!config.auth.did) {
			// 	//     config.auth.did = config.email;
			// 	// }
			// 	servers.push(new DIDAuthServer(new ConsoleMessenger()));
			// 	break;
			// case AuthType.Totp:
			// 	if (!config.auth.totp) {
			// 		config.auth.totp = new TotpAuthConfig2();
			// 	}
			// 	servers.push(new TotpAuthServer2(config.auth.totp));
			// 	break;
			// case AuthType.WebAuthnPlatform:
			// case AuthType.WebAuthnPortable:
			// 	if (servers.some((s) => s.supportsType(type))) {
			// 		continue;
			// 	}
			// 	if (!config.auth.webauthn) {
			// 		const clientHostName = new URL(config.server.clientUrl)
			// 			.hostname;
			// 		config.auth.webauthn = new WebAuthnConfig({
			// 			rpID: clientHostName,
			// 			rpName: clientHostName,
			// 			origin: config.server.clientUrl
			// 		});
			// 	}
			// 	{
			// 		const webauthServer = new WebAuthnServer(
			// 			config.auth.webauthn
			// 		);
			// 		await webauthServer.init();
			// 		servers.push(webauthServer);
			// 	}
			// 	break;
			// case AuthType.PublicKey:
			// 	servers.push(new PublicKeyAuthServer());
			// 	break;
			// case AuthType.OpenID:
			// 	servers.push(new OpenIDServer(config.auth.openid!));
			// 	break;

			// case AuthType.OsPassword:
			// 	servers.push(new OsPasswordAuthServer(bflUrl));
			// 	break;

			// case AuthType.OsToken:
			// 	servers.push(new OsTokenAuthServer(bflUrl));
			// 	break;
			case AuthType.SSI:
				servers.push(new SSIAuthServer());
				break;
			default:
				throw `Invalid authentication type: "${type}" - supported values: ${Object.values(
					AuthType
				)}`;
		}
	}
	return servers;
}

async function initProvisioner(
	config: PadlocConfig,
	storage: Storage /*, directoryProviders?: DirectoryProvider[]*/
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

// async function initDirectoryProviders(config: PadlocConfig, storage: Storage) {
//     if (!config.directory) {
//         return [];
//     }
//     let providers: DirectoryProvider[] = [];
//     for (const provider of config.directory.providers) {
//         switch (provider) {
//             case "scim":
//                 if (!config.directory.scim) {
//                     config.directory.scim = new ScimServerConfig();
//                 }
//                 const scimServer = new ScimServer(config.directory.scim, storage);
//                 await scimServer.init();
//                 providers.push(scimServer);
//                 break;
//             default:
//                 throw `Invalid value for PL_DIRECTORY_PROVIDERS: ${provider}! Supported values: "scim"`;
//         }
//     }
//     return providers;
// }

async function initChangeLogger(
	config: ChangeLogConfig,
	defaultStorage: Storage
) {
	if (!config) {
		return;
	}

	const storage = config.storage
		? await initDataStorage(config.storage)
		: defaultStorage;

	return new ChangeLogger(storage, config);
}

async function initRequestLogger(
	config: RequestLogConfig,
	defaultStorage: Storage
) {
	if (!config) {
		return;
	}

	const storage = config.storage
		? await initDataStorage(config.storage)
		: defaultStorage;

	return new RequestLogger(storage, config);
}

async function init(config: PadlocConfig) {
	setPlatform(new NodePlatform());

	const emailSender = await initEmailSender(config.email);
	const storage = await initDataStorage(config.data);
	const logger = await initLogger(config.logging);
	const attachmentStorage = await initAttachmentStorage(config.attachments);
	const authServers = await initAuthServers(config);
	//const directoryProviders = await initDirectoryProviders(config, storage);
	const provisioner = await initProvisioner(
		config,
		storage /*, directoryProviders*/
	);

	// if (config.directory.scim && !config.server.scimServerUrl) {
	//     config.server.scimServerUrl = config.directory.scim.url;
	// }

	const changeLogger = await initChangeLogger(config.changeLog, storage);
	const requestLogger = await initRequestLogger(config.requestLog, storage);

	const server = new Server(
		config.server,
		storage,
		emailSender,
		logger,
		authServers,
		attachmentStorage,
		provisioner,
		bflUrl,
		signUrl,
		authUrl,
		changeLogger,
		requestLogger
	);

	//new DirectorySync(server.makeController({ id: await uuid() }), directoryProviders);

	// Skip starting listener if --dryrun flag is present
	if (process.argv.includes('--dryrun')) {
		process.exit(0);
	}

	console.log(`Starting server on port ${config.transport.http.port}`);
	new HTTPReceiver(config.transport.http).listen((req) => server.handle(req));

	// Notify admin if any uncaught exceptions cause the program to restart
	process.on('uncaughtException', async (err: Error) => {
		console.error('uncaught exception: ', err.message, err.stack, 'exiting...');
		if (config.server.reportErrors) {
			// try {
			//     await emailSender.send(
			//         config.server.reportErrors,
			//         new PlainMessage({
			//             message: `An uncaught exception occured at ${new Date().toISOString()}:\n${err.message}\n${
			//                 err.stack
			//             }`,
			//         })
			//     );
			// } catch (e) {}
		}
		process.exit(1);
	});
}

async function start() {
	const config = getConfig();
	try {
		await init(config);
		console.log(
			'Server started with config: ',
			JSON.stringify(
				stripPropertiesRecursive(config.toRaw(), ['kind', 'version']),
				null,
				4
			)
		);
	} catch (e) {
		console.error(
			'Init failed. Error: ',
			e,
			'\nConfig: ',
			JSON.stringify(
				stripPropertiesRecursive(config.toRaw(), ['kind', 'version']),
				null,
				4
			)
		);
	}
}

start();
// function DirectoryProvider() {
//     throw new Error("Function not implemented.");
// }
