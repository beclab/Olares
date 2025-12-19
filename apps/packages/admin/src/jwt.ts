import { base58btc } from 'multiformats/bases/base58';
import { base64url } from 'multiformats/bases/base64';
import * as varint from 'varint';
import { DIDDocument, LDKeyType, PublicJwk, PrivateJwk } from '@bytetrade/core';
const ED25519_CODEC_ID = varint.encode(parseInt('0xed', 16));

export function resolve(did: string): DIDDocument {
  const [scheme, method, id] = did.split(':');

  if (scheme !== 'did') {
    throw new Error('malformed scheme');
  }

  if (method !== 'key') {
    throw new Error('did method MUST be "key"');
  }

  const idBytes = base58btc.decode(id);
  const publicKeyBytes = idBytes.slice(ED25519_CODEC_ID.length);
  const x = base64url.baseEncode(publicKeyBytes);

  console.log(
    'base64url.baseEncode(publicKeyBytes) ' +
      base64url.baseEncode(publicKeyBytes),
  );

  const mId = `#${id}`;

  const didDocument: DIDDocument = {
    '@context': [
      'https://www.w3.org/ns/did/v1',
      'https://w3id.org/security/suites/ed25519-2020/v1',
      'https://w3id.org/security/suites/x25519-2020/v1',
    ],
    id: did,
    verificationMethod: [
      {
        id: mId,
        type: LDKeyType.Ed25519VerificationKey2020,
        controller: did,
        publicKeyJwk: {
          alg: 'EdDSA',
          crv: 'Ed25519',
          kid: did,
          kty: 'OKP',
          use: 'sig',
          x: x,
        },
      },
    ],
    authentication: [mId],
    assertionMethod: [mId],
    capabilityDelegation: [mId],
    capabilityInvocation: [mId],
  };

  return didDocument;
}
