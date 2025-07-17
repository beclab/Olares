---
outline: [2, 3]
---

# Management with SDK

## DID

### Get all DIDs

Sometimes we need complete data for statistical or other reasons. The following `fetchAll` method utilizes the contract interfaces to get complete data. Although it gets the result directly from on-chain interfaces instead of traversing Ethereum events, this can be time-consuming as the amount of data is large.

```Typescript
import DID from 'did-contract-developer-components'

const RPC = "you-rpc-url"
const CONTRACT_DID = "0x5da4fa8e567d86e52ef8da860de1be8f54cae97d"
const CONTRACT_ROOT_RESOLVER = "0xe2eaba0979277a90511f8873ae1e8ca26b54e740"
const CONTRACT_ABI_TYPE = "0x9ae3f16bd99294af1784beb1a0a5c84bf2636365"

const fetch = async () => {
	const did = DID.createConsole(RPC, CONTRACT_DID, CONTRACT_ROOT_RESOLVER, CONTRACT_ABI_TYPE)

	const dids = await did.fetchAll()

	console.log('dids:', dids)

	console.log('format dids:', await did.formatDatas(dids))
}

fetch()
```

> [!NOTE]
> If running in development environment and logging is on, the fetching progress is shown during execution.

#### Fast fetch

For your convenience, the SDK provides two functions `formatDatas` and `loadDatas`. The data returned by `formatDatas` can be stored and it can be loaded by `loadDatas` next time to reduce syncing duration. A simpler way is to access the `/all` endpoint in the official **did-support** service, which can also be loaded by `loadDatas`.

### Query specific DID

After reading the above contract interfaces, you should notice that the contract does not provide a single interface for fetching complete data related to a DID. So we simplified the design of interfaces in the SDK, where `fetchDomain` returns complete data for a DID.

```Typescript
import DID from 'did-contract-developer-components'

const RPC = "you-rpc-url"
const CONTRACT_DID = "0x5da4fa8e567d86e52ef8da860de1be8f54cae97d"
const CONTRACT_ROOT_RESOLVER = "0xe2eaba0979277a90511f8873ae1e8ca26b54e740"
const CONTRACT_ABI_TYPE = "0x9ae3f16bd99294af1784beb1a0a5c84bf2636365"

const fetch = async () => {
	const did = DID.createConsole(RPC, CONTRACT_DID, CONTRACT_ROOT_RESOLVER, CONTRACT_ABI_TYPE)

	const domain = await did.fetchDomain('james.myterminus.com')

	console.log('did:', domain)
}

fetch()
```

#### Fast query

After you execute `loadDatas` in the above SDK examples, `fetchDomain` will first try to match local data and only fetch data on-chain if there is no data for this DID on local machine.

#### Update DID

If you are worried that fast query does not return the latest data, use `updateDomain` or `updateDomainById` to update local data.

#### Fuzzy matching

Like fast query, you can use the fuzzy matching function in SDK after loading data locally. The following two functions only match the local data:

```Typescript
import DID from 'did-contract-developer-components'
//... code

// Query DID by owner and return all subdomains
const domainsByOwner = DID.Domain.findSubtreesByOwner(owner, did.treesCache)

//Query DID by did and return all subdomains
const domainsByDid = DID.Domain.findSubtreesByDid(did, did.treesCache)

```

## Tag

### Get all tags of a DID

Unlike calling the contract directly, this function returns data with parsed tag structure and struct field names.

```Typescript
import DID from 'did-contract-developer-components'

const RPC = "you-rpc-url"
const CONTRACT_DID = "0x5da4fa8e567d86e52ef8da860de1be8f54cae97d"
const CONTRACT_ROOT_RESOLVER = "0xe2eaba0979277a90511f8873ae1e8ca26b54e740"
const CONTRACT_ABI_TYPE = "0x9ae3f16bd99294af1784beb1a0a5c84bf2636365"

const fetch = async () => {
	const did = DID.createConsole(RPC, CONTRACT_DID, CONTRACT_ROOT_RESOLVER, CONTRACT_ABI_TYPE)

	const domain = await did.fetchDomain('james.myterminus.com')

	if (domain == undefined) {
		throw new Error("not found");
	}

	const tags = await DID.Domain.fetchAllTagType(domain, did.getContractDID())

	console.log('tags:', tags)
}

fetch()
```

### Get all tag values of a DID

Like the above method, this function returns parsed data with JSON tag values.

```Typescript
import DID from 'did-contract-developer-components'

const RPC = "you-rpc-url"
const CONTRACT_DID = "0x5da4fa8e567d86e52ef8da860de1be8f54cae97d"
const CONTRACT_ROOT_RESOLVER = "0xe2eaba0979277a90511f8873ae1e8ca26b54e740"
const CONTRACT_ABI_TYPE = "0x9ae3f16bd99294af1784beb1a0a5c84bf2636365"

const fetch = async () => {
	const did = DID.createConsole(RPC, CONTRACT_DID, CONTRACT_ROOT_RESOLVER, CONTRACT_ABI_TYPE)

	const domain = await did.fetchDomain('james.myterminus.com')

	if (domain == undefined) {
		throw new Error("not found");
	}

	await DID.Domain.fetchAllTagType(domain, did.getContractDID())

	const tags = await DID.Domain.fetchAllTagValue(domain, did.getContractDID())

	console.dir(tags, {depth: null});
}

fetch()
```

### Define tag

Using SDK to define a tag, you can construct the inner structure of the tag type with an object-oriented approach without worrying about the encoding. The following examples include most common data types.

```Typescript
import DID from 'did-contract-developer-components'

const RPC = "you-rpc-url"
const CONTRACT_DID = "0x5da4fa8e567d86e52ef8da860de1be8f54cae97d"
const CONTRACT_ROOT_RESOLVER = "0xe2eaba0979277a90511f8873ae1e8ca26b54e740"
const CONTRACT_ABI_TYPE = "0x9ae3f16bd99294af1784beb1a0a5c84bf2636365"

const defineTag = async () => {

	// tag name: simpleTag
	const tagName = 'simpleTagBox'

	// tuple
	const testTuple = new DID.Tag.TagValueTypeTuple(undefined, undefined, true);

	// uint
	const testUint = new DID.Tag.TagValueTypeUint(undefined, undefined, true);
	testUint.setSize(8)

	// address
	const testAddress = new DID.Tag.TagValueTypeAddress(undefined, undefined, true);

	// array<address>
	const testArrayAddress = new DID.Tag.TagValueTypeArray(undefined, undefined, true);
	testArrayAddress.setBuliderType(testAddress)

	// bool
	const testBool= new DID.Tag.TagValueTypeBool(undefined, undefined, true)

	// bytes
	const testBytes = new DID.Tag.TagValueTypeBytes(undefined, undefined, true)

	// int
	const testInt = new DID.Tag.TagValueTypeInt(undefined, undefined, true)
	testInt.setSize(256)

	// flarray<int>
	const testFlarrayInt = new DID.Tag.TagValueTypeFlarray(undefined, undefined, true)
	testFlarrayInt.setBuliderType(testInt)
	testFlarrayInt.setSize(3)

	// flbytes
	const testFlbytes = new DID.Tag.TagValueTypeFlbytes(undefined, undefined, true)
	testFlbytes.setSize(5)

	// string
	const testString = new DID.Tag.TagValueTypeString(undefined, undefined, true)

	testTuple.setField('testUint', testUint)
	testTuple.setField('testArrayAddress', testArrayAddress)
	testTuple.setField('testBool', testBool)
	testTuple.setField('testBytes', testBytes)
	testTuple.setField('testFlarrayInt', testFlarrayInt)
	testTuple.setField('testFlbytes', testFlbytes)
	testTuple.setField('testString', testString)

	const did = DID.createConsole(RPC, CONTRACT_DID, CONTRACT_ROOT_RESOLVER, CONTRACT_ABI_TYPE)

	const domain = await did.fetchDomain('james.myterminus.com')

	await DID.Domain.defineTag(domain, tagName, testTuple, did.getContractDIDByPrivateKey('you-private-key'), did)
}

defineTag()
```

### Set tagger

```Typescript
import DID from 'did-contract-developer-components'

const RPC = "you-rpc-url"
const CONTRACT_DID = "0x5da4fa8e567d86e52ef8da860de1be8f54cae97d"
const CONTRACT_ROOT_RESOLVER = "0xe2eaba0979277a90511f8873ae1e8ca26b54e740"
const CONTRACT_ABI_TYPE = "0x9ae3f16bd99294af1784beb1a0a5c84bf2636365"

const setTagger = async () => {

	const did = DID.createConsole(RPC, CONTRACT_DID, CONTRACT_ROOT_RESOLVER, CONTRACT_ABI_TYPE)

	const domain = await did.fetchDomain('james.myterminus.com')

	if (domain == undefined) {
		throw new Error("not found");
	}

	const [tag] = domain.tags.filter(tag => tag.name == 'simpleTagBox')

	DID.Domain.setTagger(tag, 'you-tagger-address', did.getContractDIDByPrivateKey('you-private-key'))
}

setTagger()
```

### Set tag value

Like defining a tag above, when using SDK to set tag value, you only need to pass in suitable JSON data without worrying about the encoding. The tag types in the following examples are the same as the tag definition section above.

```Typescript
import DID from 'did-contract-developer-components'

const RPC = "you-rpc-url"
const CONTRACT_DID = "0x5da4fa8e567d86e52ef8da860de1be8f54cae97d"
const CONTRACT_ROOT_RESOLVER = "0xe2eaba0979277a90511f8873ae1e8ca26b54e740"
const CONTRACT_ABI_TYPE = "0x9ae3f16bd99294af1784beb1a0a5c84bf2636365"

const setTagValue = async () => {
	const did = DID.createConsole(RPC, CONTRACT_DID, CONTRACT_ROOT_RESOLVER, CONTRACT_ABI_TYPE)

	const domain = await did.fetchDomain('james.myterminus.com')

	if (domain == undefined) {
		throw new Error("not found");
	}

	console.log(domain)
	const [tag] = domain.tags.filter(tag => tag.name == 'simpleTagBox')
	const tagType = await DID.Domain.fetchTagStructure(tag, did.getContractDID())
	console.dir(tagType, {depth: null})

	const ba1 = ethers.hexlify(Uint8Array.from([1, 2, 3]))
	const ba2 = ethers.hexlify(Uint8Array.from([10, 18, 19]))

	const newData = {
		testbox: {
			testUint: 1,
			testArrayAddress: ['0xF18B2Ea28c722CA87f951F5bF5327b66a7dd72A3', '0xecBA1d33b889f66ad426535f970d1E033ba5c79C'],
			testBool: true,
			testBytes: '0x0102030405' ,
			testFlarrayInt: [2, 3, 4],
			testFlbytes: '0x0a0b0c0d0e',
			testString: 'ok'
		}
	}
	console.log('newData', newData)

	const call = await DID.Tag.doEncode(tagType, newData)
	console.log('call', call)

	//set
	const resp = await DID.Domain.setValue(tag, call, did.getContractDIDByPrivateKey('you-private-key'))
}

setTagValue()
```
