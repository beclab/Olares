---
outline: [2, 3]
---

# 直接调用合约

## DID

### 获取元数据 (metadata)

获取 DID 的元数据有两个接口 - `getMetadata(tokenId)` 和 `getMetadata(name)`。

#### 返回值

```json
[
  "james.myterminus.com", // 名称 (name)
  "did:key:z6MkpLwxcTwhj4MRm4eKhvBadK45qHr5QEYHUXNyhCfkXJ9U#z6MkpLwxcTwhj4MRm4eKhvBadK45qHr5QEYHUXNyhCfkXJ9U", // 从助记词派生的 DID
  "OrganizationalUser", // DID 类型
  true // 是否允许创建子域？
]
````

### 获取域的所有者

调用 `ownerOf(tokenId)` 来获取 DID 的链上控制者地址。

### 按创建索引获取 token

调用 `tokenByIndex(index)` 来获取指定索引的 token。它返回一个 **token ID**。

### 按所有者和索引获取 token

尽管从业务角度我们不允许单个钱包拥有多个 DID，但考虑到未来可能的所有权转移和 NFT 交易，合约本身是允许的。在这种情况下，可以调用 `tokenOfOwnerByIndex(owner, index)` 来获取指定地址拥有的、具有指定索引的 token。它返回一个 **token ID**。

### 注册 DID

域的所有者可以调用 `register(owner, MetaData(domain, did, note, allowSubdomain))` 来注册其子域。

> [\!NOTE]
> 第一个参数 **owner** 是新 DID 的指定所有者，第二个参数是一个元数据结构体，包含：
>
>   - domain: 新 DID 的完整域名，同时也是一个 Terminus Name
>   - did: 从所有者钱包派生的 DID
>   - note: 关于新 DID 的备注，供链下系统用于分类
>   - allowSubdomain: 是否允许新 DID 注册子域
>
> 元数据在注册后不可更改。如果将来所有权发生转移，新的 DID 记录将被写入 `latestDid` 标签中。

## 标签 (Tag)

### 获取由某个名称定义的标签数量

`getDefinedTagCount(name)` 返回由指定 TName 定义的标签数量。

### 获取标签名

与上述接口配合使用，`getDefinedTagNameByIndex(name, index)` 返回单个标签名，而 `getDefinedTagNames(name)` 返回由指定 TName 定义的所有标签名。

### 结构化标签

如果一个标签类型是复杂结构而不是原始值，调用 `getTagType(name, tagName)` 来查询结构定义，然后使用先前返回的 **fieldNamesHash** 调用 `getFieldNamesEventBlock(fieldNamesHash)` 来获取定义此标签时的区块号。最后，使用 `ethers` 库来获取定义中的字段名。

> [\!NOTE]
> 查询标签类型的接口返回一个 ABI 类型的编码字节，应根据代码表进行解析。查询字段名可能很复杂且容易出错，因此我们建议使用 SDK 中的函数来获取关于标签的数据，而不是手动调用合约。

获取标签类型后，调用 `getTagElem(definedDidName, valueDidName, tagName, elemPath)` 来获取标签的值。

> [\!NOTE]
>
>   - definedDidName: 定义此标签的 TName
>   - valueDidName: 你想获取其标签值的 TName
>   - elemPath: 路径选择器，用于读取元组或数组中的内部元素（`[]` 用于读取完整标签）
>
> 此接口返回 ABI 编码的数据，应根据标签类型进行解析。
> 同样建议使用 SDK 而不是手动调用合约。

### 定义标签

使用 `didName` 的所有者调用 `defineTag(didName, tagName, abiType, fieldNames)` 来定义一个标签。

> [\!NOTE]
>
>   - abiType: 此标签的数据类型，遵循我们的 **ABI** 代码格式，支持复杂结构
>   - fieldNames: 此标签内部结构体的字段名，使用前序遍历扁平化为二维字符串数组

### CRUD (增删改查)

> [\!NOTE]
> 我们建议与 Tagger 合约交互，而不是使用以下接口来执行标签操作。

- **创建 (Create)** - 调用 `addTag(defineDidName, valueDidName, tagName, value)` 来添加一个标签。

- **更新 (Update)** - 调用 `updateTagElem(defineDidName, valueDidName, tagName, elemPath, value)` 来更新标签中的一部分数据。

- **删除 (Delete)** - 调用 `removeTag(defineDidName, valueDidName, tagName)` 来删除一个标签。这会删除在某个 DID 上设置的标签，而不是删除标签定义本身。

- **数组操作 (Array operations)** - 如果标签类型包含数组，调用 `popTagElem(defineDidName, valueDidName, tagName, elePaths)` 和 `pushTagElem(defineDidName, valueDidName, tagName, elePaths, value)` 来执行数组特有的操作。

> [\!NOTE]
>
>   - defineDidName: 定义此标签的 TName
>   - valueDidName: 在其上设置标签值的 TName
>   - value: ABI 编码的标签值
