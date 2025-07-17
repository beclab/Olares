---
description: Olares 专业术语词汇表，提供系统中常见技术概念的标准定义和解释。按字母顺序排列，便于查找和理解 Olares 生态中的专业术语。
---
# Glossary
<div style="text-align: center; font-size: 18px; margin-bottom: 20px;">
  <a href="#c" style="text-decoration: none; color: #007BFF;">C</a>
  <span style="margin: 0 10px;">|</span>
  <a href="#d" style="text-decoration: none; color: #007BFF;">D</a>
  <span style="margin: 0 10px;">|</span>
  <a href="#f" style="text-decoration: none; color: #007BFF;">F</a>
  <span style="margin: 0 10px;">|</span>
  <a href="#t" style="text-decoration: none; color: #007BFF;">T</a>
  <span style="margin: 0 10px;">|</span>
  <a href="#v" style="text-decoration: none; color: #007BFF;">V</a>
</div>

## C
### CNAME Record
CNAME（规范名称）记录是 DNS（域名系统）记录的一种，用于将自定义域名映射到 Olares 提供的地址，从而实现域名到应用的映射。

## D
### DID
去中心化标识符（DID）是一种独特的数字身份识别方法，可以让任何个人或实体在网络上拥有持久且唯一的身份标识。这种标识符完全独立，无需通过中心化机构进行验证或注册。DID 是一个以“did”开头的特定格式字符串，遵循既定的格式和命名规范。

## F
### FRP
快速反向代理（FRP）是一款专为内网穿透场景设计的高性能反向代理工具。它使位于 NAT 或防火墙后的服务器即使没有公网 IP，也能对外提供服务。通过 FRP，用户可以轻松地将内部服务暴露到公网。

## T
### TOTP
基于时间的一次性密码（TOTP）是一种用于生成一次性密码的时间相关算法，在双因素认证（2FA）中广泛使用。它通过共享密钥和当前时间生成一次性密码，以加强账户安全性。每个密码仅在短时间内（通常是 30 或 60 秒）有效，之后系统会生成新的密码。

## V
### VC
可验证凭证（VC）是一种数字化的证明格式，可以验证持有者的特定属性或资质，同时不会泄露额外的个人信息。它涉及三个主要角色：

* 持有者：拥有并使用凭证来证明特定信息的个人
* 颁发者：创建和签发可验证凭证的权威机构或实体
* 验证者：需要通过验证凭证来确认持有者信息真实性的个人或组织