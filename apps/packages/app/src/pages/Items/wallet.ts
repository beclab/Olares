// import { i18n } from '../../boot/i18n';

export class Wallet {
	// importMnemonic = async (name: string, mnemonic: string, isImport = true) => {
	// 	if (!name) {
	// 		throw new Error(i18n.global.t('errors.invalid_name'));
	// }
	// if (mnemonic || isImport) {
	// }
	// if (!supportCenter.walletCore.Mnemonic.isValid(mnemonic)) {
	//     throw new Error(
	//         i18n.global.t('The seed phrase is invalid, please check!')
	//     );
	// }
	// const sameMnemonicKeyring = this.getKeyringByMnemonic(mnemonic);
	// if (sameMnemonicKeyring) {
	//     throw new Error(
	//         i18n.global.t('The seed has been imported, please check!')
	//     );
	// }
	// const pk58 = nearProtocol.mnemonicToPublic58(mnemonic);
	// const nearAccountId = await walletController.getNearAccountId(pk58, true);
	// const keyring = await keyringService.importMnemonics(
	//     name,
	//     mnemonic,
	//     isImport,
	//     this.generateKeyringId(),
	//     nearAccountId
	// );
	// if (!keyring) {
	//     throw new Error($l('keyring creation failed!'));
	// }
	// const displayKeyring = await keyringService.getKeyringById(
	//     keyring.keyringId
	// );
	// if (displayKeyring) {
	//     assetService.keyringAssetInit(displayKeyring);
	// }
	// preferenceService.setCurrentKeyringId(keyring.keyringId);
	// return this.foregroundGetKeyringById();
	// };
	// importPrivateKey = async (name: string, data: string) => {
	// 	if (!name) {
	// 		throw new Error(i18n.global.t('errors.invalid_name'));
	// 	}
	// if (data) {
	// }
	// if (supportCenter.walletCore.PrivateKey.isValid(Buffer.from(data), supportCenter.walletCore.Curve.secp256k1)) {
	//     throw new Error(
	//         i18n.global.t('the private key is invalid')
	//     );
	// }
	// let chain = CHAINS_ENUM.NONE;
	// const pk58 = nearProtocol.privateToPublic58(data);
	// const nearAccountId = await walletController.getNearAccountId(pk58, true);
	// if (nearAccountId) {
	//     chain = CHAINS_ENUM.NEAR_PROTOCOL;
	// }
	// const keyring = await keyringService.importPrivateKey(
	//     name,
	//     data,
	//     this.generateKeyringId(),
	//     chain,
	//     nearAccountId
	// );
	// if (!keyring) {
	//     throw new Error(i18n.global.t('keyring creation failed!'));
	// }
	// const displayKeyring = await keyringService.getKeyringById(
	//     keyring.keyringId
	// );
	// if (displayKeyring) {
	//     assetService.keyringAssetInit(displayKeyring);
	// }
	// preferenceService.setCurrentKeyringId(keyring.keyringId);
	// return this.foregroundGetKeyringById();
	// };
	// importKeyStore = async (name: string, content: string, password: string) => {
	// 	if (!name) {
	// 		throw new Error(i18n.global.t('errors.invalid_name'));
	// 	}
	// 	if (content === undefined || password === undefined) {
	// 	}
	// try {
	//     JSON.parse(content);
	// } catch {
	//     throw new Error(i18n.global.t('the input file is invalid'));
	// }
	// const { StoredKey, HexCoding } = supportCenter.walletCore;
	// const storeKey = StoredKey.importJSON(Buffer.from(content))
	// const passwordBuffer = Buffer.from(password);
	// if (storeKey.isMnemonic()) {
	//     const mnemonic = storeKey.decryptMnemonic(passwordBuffer)
	//     storeKey.delete();
	//     return await this.importMnemonic(name, mnemonic)
	// } else {
	//     const keyBuffer = storeKey.decryptPrivateKey(passwordBuffer);
	//     const key = HexCoding.encode(keyBuffer).replace('0x', '');
	//     storeKey.delete();
	//     return await this.importPrivateKey(name, key)
	// }
	// };
}

export default new Wallet();
