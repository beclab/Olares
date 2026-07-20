---
outline: [2, 3]
---

# 使用 SDK 进行管理

## DID

### 获取所有 DID

有时出于统计或其他原因，我们需要完整的数据。以下 `fetchAll` 方法利用合约接口来获取完整数据。虽然它直接从链上接口获取结果，而不是遍历以太坊事件，但由于数据量大，这可能会非常耗时。

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
````

> [\!NOTE]
> 如果在开发环境中运行且日志记录已开启，执行期间会显示获取进度。

#### 快速获取

为方便起见，SDK 提供了 `formatDatas` 和 `loadDatas` 两个函数。`formatDatas` 返回的数据可以被存储，下次通过 `loadDatas` 加载以减少同步时间。一个更简单的方法是访问官方 **did-support** 服务中的 `/all` 端点，其返回的数据也可以被 `loadDatas` 加载。

### 查询特定 DID

阅读了上述合约接口后，你应该会注意到合约并未提供单个接口来获取与某个 DID 相关的完整数据。因此，我们简化了 SDK 中的接口设计，其中 `fetchDomain` 会返回一个 DID 的完整数据。

```typescript
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

#### 快速查询

在你执行了上述 SDK 示例中的 `loadDatas` 后，`fetchDomain` 会首先尝试匹配本地数据，仅当本地机器上没有此 DID 的数据时，才会从链上获取。

#### 更新 DID

如果你担心快速查询返回的不是最新数据，可以使用 `updateDomain` 或 `updateDomainById` 来更新本地数据。

#### 模糊匹配

与快速查询类似，在本地加载数据后，你可以使用 SDK 中的模糊匹配功能。以下两个函数仅匹配本地数据：

```typescript
import DID from 'did-contract-developer-components'
//... code

// 按 owner 查询 DID 并返回所有子域
const domainsByOwner = DID.Domain.findSubtreesByOwner(owner, did.treesCache)

// 按 did 查询 DID 并返回所有子域
const domainsByDid = DID.Domain.findSubtreesByDid(did, did.treesCache)

```

## 标签 (Tag)

### 获取某个 DID 的所有标签

与直接调用合约不同，此函数返回的数据中包含了已解析的标签结构和结构体字段名。

```typescript
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

### 获取某个 DID 的所有标签值

与上述方法类似，此函数返回带有 JSON 格式标签值的已解析数据。

```typescript
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

### 定义标签

使用 SDK 定义标签时，你可以使用面向对象的方法构建标签类型的内部结构，而无需担心编码问题。以下示例包含了大多数常见的数据类型。

```typescript
import DID from 'did-contract-developer-components'

const RPC = "you-rpc-url"
const CONTRACT_DID = "0x5da4fa8e567d86e52ef8da860de1be8f54cae97d"
const CONTRACT_ROOT_RESOLVER = "0xe2eaba0979277a90511f8873ae1e8ca26b54e740"
const CONTRACT_ABI_TYPE = "0x9ae3f16bd99294af1784beb1a0a5c84bf2636365"

const defineTag = async () => {

    // 标签名: simpleTagBox
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

### 设置 Tagger

```typescript
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

### 设置标签值

与上面定义标签类似，当使用 SDK 设置标签值时，你只需传入合适的 JSON 数据，无需担心编码问题。以下示例中的标签类型与上面定义标签部分中的相同。

```typescript
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

    //设置
    const resp = await DID.Domain.setValue(tag, call, did.getContractDIDByPrivateKey('you-private-key'))
}

setTagValue()
```
