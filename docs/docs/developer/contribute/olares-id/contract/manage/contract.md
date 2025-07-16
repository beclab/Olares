---
outline: [2, 3]
---

# Call the contract directly

## DID

### Get metadata

There are two interfaces for fetching the metadata of a DID - `getMetadata(tokenId)` and `getMetadata(name)`.

#### Return value

```json
[
  "james.myterminus.com", // name
  "did:key:z6MkpLwxcTwhj4MRm4eKhvBadK45qHr5QEYHUXNyhCfkXJ9U#z6MkpLwxcTwhj4MRm4eKhvBadK45qHr5QEYHUXNyhCfkXJ9U", // DID derived from mnemonic phrases
  "OrganizationalUser", // DID type
  true // allowed to create subdomains?
]
```

### Get the owner of a domain

Call `ownerOf(tokenId)` to get the on-chain controller address of a DID.

### Get token by index of creation

Call `tokenByIndex(index)` to get the token with a specified index. It returns a **token ID**.

### Get token by owner and index

Although we disallow owning multiple DIDs by a single wallet from the business' perspective, the contract allows this considering possible ownership transferring and NFT trading in the future. In this case, `tokenOfOwnerByIndex(owner, index)` can be called to get the token owned by a specified address with a specified index. It returns a **token ID**.

### Register DID

The owner of a domain can call `register(owner, MetaData(domain, did, note, allowSubdomain))` to register its subdomains.

> [!NOTE]
> The first parameter **owner** is the specified owner of the new DID and the second parameter is a struct of metadata containing.
>
> - domain: the complete domain name of the new DID, which is also a Olares ID
> - did: the DID derived from the owner's wallet
> - note: notes about the new DID, used by off-chain systems for categorization
> - allowSubdomain: whether to allow the new DID to register subdomains
>
> The metadata cannot be changed after registration. If the ownership is transferred in the future, the new DID record will be written to the `latestDid` tag.

## Tag

### Get the number of tags defined by a name

`getDefinedTagCount(name)` returns how many tags are defined by a specified TName.

### Get tag name

Used with the above interface, `getDefinedTagNameByIndex(name, index)` returns a single tag name and `getDefinedTagNames(name)` returns all tag names defined by a specified TName.

### Structured tag

If a tag type is a complicated structure instead of primitive value, call `getTagType(name, tagName)` to query the structure definition and then call `getFieldNamesEventBlock(fieldNamesHash)` with previously returned **fieldNamesHash** to get the block number at which this tag is defined. Finally, use the `ethers` library to get field names in the definition.

> [!NOTE]
> The interface for querying tag type returns an encoded bytes of the ABI type, which should be parsed according to the code table. Querying field names can be complicated and error-prone, so we recommend to use functions in the SDK to fetch data about tags instead of calling the contract manually.

Call `getTagElem(definedDidName, valueDidName, tagName, elemPath)` to get the value of a tag after getting the tag type.

> [!NOTE]
>
> - definedDidName: the TName defining this tag
> - valueDidName: the TName whose tag value is desired
> - elemPath: path selector to read an inner element for tuples or arrays (`[]` for reading the full tag)
>
> This interface returns ABI-encoded data which should be parsed according to the tag type.
> It is also recommended to use SDK instead of calling the contract manually.

### Define a tag

Call `defineTag(didName, tagName, abiType, fieldNames)` with the owner of `didName` to define a tag.

> [!NOTE]
>
> - abiType: data type of this tag following our **ABI** code format, which supports complicated structures
> - fieldNames: the field names of structs inside this tag, flatten as a 2D string array using pre-order traversal

### CRUD

> [!NOTE]
> We recommend to interact with the tagger contract instead of using the following interfaces to perform tag operations.

- Create - Call `addTag(defineDidName, valueDidName, tagName, value)` to add a tag.

- Update - Call `updateTagElem(defineDidName, valueDidName, tagName, elemPath, value)` to update a piece of data in a tag.

- Delete - Call `removeTag(defineDidName, valueDidName, tagName)` to delete a tag. This deletes a tag set on a DID instead of deleting the tag definition.

- Array operations - If the tag type contains an array, call `popTagElem(defineDidName, valueDidName, tagName, elePaths)` and `pushTagElem(defineDidName, valueDidName, tagName, elePaths, value)` to perform array-specific operations.

> [!NOTE]
>
> - defineDidName: the TName defining this tag
> - valueDidName: the TName on which the tag value is set
> - value: ABI-encoded tag value
