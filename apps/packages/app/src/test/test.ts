// import {WebCryptoProvider} from '../lib/crypto'
// import {WebCryptoProviderLocal} from '../lib/localcrypto'
// //import { HashParams } from '@bytetrade/core';
// import { AESKeyParams, PBKDF2Params, AESEncryptionParams, RSAPublicKeyParams,RSAKeyParams, RSAEncryptionParams, RSASigningParams , HMACKeyParams,HMACParams } from '@bytetrade/core';
// import {  stringToBytes, } from "@didvault/core/src/encoding";

//  const wp = new WebCryptoProvider();
//  const wpl = new WebCryptoProviderLocal();

// console.log(await wp.randomBytes(25))
// console.log(await wpl.randomBytes(25))

// const NodeRSA = require('node-rsa');
// const key = new NodeRSA({b: 512});

// const text = 'Hello RSA!';
// const encrypted = key.encrypt(text, 'base64');
// console.log(encrypted)

// console.log(await wp.hash(new Uint8Array([1]), new HashParams( {algorithm: 'SHA-1' } )))
// console.log(await wpl.hash(new Uint8Array([1]),new HashParams( {algorithm: 'SHA-1' } )))

// console.log(await wp.hash(new Uint8Array([1]), new HashParams( {algorithm: 'SHA-256' } )))
// console.log(await wpl.hash(new Uint8Array([1]),new HashParams( {algorithm: 'SHA-256' } )))

// let password = '123456';
// let keyParams:PBKDF2Params = new PBKDF2Params();
// keyParams.salt = await wp.randomBytes(16);

// //DeriveKey
// console.log(await wp.deriveKey(stringToBytes(password), keyParams));
// console.log(await wpl.deriveKey(stringToBytes(password), keyParams));

// //AES encryption
// let encryptionParams = new AESEncryptionParams();
// let key = await wpl.generateKey(new AESKeyParams());
// let data = await wpl.randomBytes(25);
// encryptionParams.iv = await wpl.randomBytes(16);
// encryptionParams.additionalData = await wpl.randomBytes(16);
// //encryptionParams.algorithm === "AES-CCM"

// let d1 = await wp.encrypt(key, data,encryptionParams);
// let d2 = await wpl.encrypt(key, data,encryptionParams);
// console.log(await wp.decrypt(key,d1,encryptionParams));
// console.log(await wpl.decrypt(key,d2,encryptionParams));

// //RSA
// const rsa1 = await wp.generateKey(new RSAKeyParams());
// //const rsa2 = await wpl.generateKey(new RSAKeyParams());
// const rsa = rsa1;

// // console.log(await wp.fingerprint(rsa1.publicKey));
// // console.log(await wpl.fingerprint(rsa2.publicKey));
// // console.log(rsa1);
// // console.log(rsa2);

// let data = await wpl.generateKey(new AESKeyParams());
// console.log(data);

// //encrypt/decrypt
// // let d1 = await wp.encrypt(rsa.publicKey, data, new RSAEncryptionParams());
// // let d2 = await wpl.encrypt(rsa.publicKey, data, new RSAEncryptionParams());
// // console.log(await wp.decrypt(rsa.privateKey,d1, new RSAEncryptionParams()));
// // console.log(await wpl.decrypt(rsa.privateKey,d1, new RSAEncryptionParams()));

// //sign/verify
// let d1 = await wp.sign(rsa.privateKey, data, new RSASigningParams());
// let d2 = await wpl.sign(rsa.privateKey, data, new RSASigningParams());
// console.log(await wpl.verify(rsa.publicKey, d1, data, new RSASigningParams()));
// console.log(await wp.verify(rsa.publicKey, d2, data, new RSASigningParams()));
// console.log(await wpl.verify(rsa.publicKey, d2, data, new RSASigningParams()));

// HMAC sign verify
// const hamckey = await wp.generateKey(new HMACKeyParams());
// let data = await wpl.generateKey(new AESKeyParams());
// let sd1 = await wp.sign(hamckey, data, new HMACParams());
// let sd2 = await wpl.sign(hamckey, data, new HMACParams());
// console.log(await wp.verify(hamckey, sd2, data, new HMACParams()));
// console.log(await wpl.verify(hamckey, sd1, data, new HMACParams()));
// console.log(await wpl.verify(hamckey, sd2, data, new HMACParams()));
