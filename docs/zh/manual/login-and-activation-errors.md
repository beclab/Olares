---
outline: [2, 3]
description: 查询 Olares 登录与激活错误信息，了解提示含义并选择下一步操作。
head:
  - - meta
    - name: keywords
      content: Olares, 登录错误, 激活错误, Authentication failed, MFA binding error, DID binding error, invalid jws
---

# 登录与激活错误信息

激活或登录 Olares 时出现明确报错，可以在本页按完整错误信息查找。实际提示与本页不同时，请原样记录，不要直接套用相似报错的原因。

## 激活错误

### `MFA binding error`

绑定多因素认证（MFA）的请求超时。检查运行 LarePass 的设备和 Olares 设备是否都能稳定访问互联网，然后重试当前激活步骤。

### `DID binding error`

激活期间访问身份绑定服务的请求超时。检查两台设备的网络连接，然后重试。

### `Invalid jws, timestamp is out of range`

运行 LarePass 的设备与 Olares 主机时间差超出允许范围。为两台设备开启自动设置日期和时间，确认时区和时间正确，然后重试。

### `Resolve name error`

Olares 主机无法解析或访问身份服务。检查互联网连接和 DNS 配置后重试。如果错误仍然出现，请先记录发生时间和当前 DNS 服务器，再收集诊断信息。

## 登录错误

### `Authentication failed, incorrect password`

输入的密码与账户不匹配。检查大小写和输入内容。如果已经忘记密码，请参考[忘记桌面登录密码](help/ts-forget-login-password.md)。

### `Authentication failed, user not found`

Olares 找不到输入的用户名。检查 Olares ID 的用户名和域名。团队成员可以请管理员前往**设置** > **用户**，确认账户是否存在。

### `Authentication failed, failed to query user from lldap service`

Olares 无法从内部身份服务读取用户记录。先重试一次。如果问题仍然存在，请管理员确认用户是否存在，记录尝试登录的时间，并[收集诊断信息](collect-diagnostic-information.md)。

### `too many failed login attempts, retry again later after 5 minutes`

连续登录失败后，Olares 暂时限制了登录。停止重试，等待至少五分钟，再输入已确认的密码。

### `Authentication failed, disk space is full`

系统磁盘没有可用空间，身份验证服务无法完成登录。请先[清理磁盘空间](free-up-disk-space.md)，然后重试。

### `Authentication failed, lldap service is unavailable`

内部身份服务当前不可用。请管理员确认 LarePass 是否同时显示**系统错误**，然后参考 [LarePass 显示“系统错误”](help/ts-system-error.md)。

### `Authentication failed, citus service is unavailable`

内部数据库服务当前不可用。请管理员检查系统状态，然后参考 [LarePass 显示“系统错误”](help/ts-system-error.md)。

## 未找到对应错误

记录完整错误信息、发生时间和时区、当前 Olares 版本，以及错误出现在激活还是登录阶段，然后[收集诊断信息](collect-diagnostic-information.md)。不要在公开 Issue 中发布完整日志、Olares ID、IP 地址、域名或账户信息。
