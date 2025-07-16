# Official Taggers

## RootTagger

So far we have defined the following tags and use `RootTagger` as their taggers:

- rsaPubKey: the RSA public key of a TName
  type: `bytes`; access: the operator, the owner of the TName, and the owner of a parent TName
- dnsARecord: the IP address of a TName
  type: `bytes4`; access: the operator, the owner of the TName, and the owner of a parent TName
- latestDID: the latest DID of a TName (this is added because the metadata is immutable)
  type: `string`; access: the operator, the owner of the TName, and the owner of a parent TName
- authAddresses: the addresses controlled by the owner of a TName
  type: `tuple(uint8,address)[]`; access: anyone with EIP-712 signatures of the owner of the TName and the added address

## AppStoreReputation

There is another special tagger `AppStoreReputation` for the tag `ratings` with type `tuple(string,uint8)[]` defined in the TName `app.myterminus.com`. It provides on-chain storage for ratings of apps in Terminus OS. Each app has a corresponding sub-TName `<appVersion>.<appId>.app.myterminus.com` where anyone who has a TName can submit ratings for these apps.
