# lxdapi API 文档

> 本文件由 `docs/swagger.json` 自动转换生成，仅供阅读参考。

## 概述

- API 路径数量: **87**
- Swagger 版本: 2.0

## Admin API - 品牌设置

### 获取品牌设置

`GET` `/api/admin/brand-settings`

获取系统品牌自定义设置

- **认证**: SessionAuth

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 500 | 获取失败 | lxdapi_pkg_response.Response |

---

### 更新品牌设置

`POST` `/api/admin/brand-settings`

更新系统品牌自定义设置，reset=true时重置为默认

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| reset | query | string | 否 | 是否重置为默认值 |
| request | body | models.BrandSettings | 否 | 品牌设置 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 更新成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 500 | 更新失败 | lxdapi_pkg_response.Response |

---

## Admin API - 缓存管理

### 获取容器缓存

`GET` `/api/admin/cache/containers`

从缓存获取所有容器信息

- **认证**: SessionAuth

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |

---

### 刷新缓存

`POST` `/api/admin/cache/refresh`

刷新容器缓存，无name参数刷新所有，有name参数刷新指定容器

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | query | string | 否 | 容器名称，不传则刷新所有 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 刷新成功 | lxdapi_pkg_response.Response |
| 500 | 刷新失败 | lxdapi_pkg_response.Response |

---

## Admin API - 认证

### 获取登录验证码

`GET` `/api/admin/captcha`

生成图片验证码用于管理员登录

- **认证**: 无

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 生成成功 | {captcha_id: string, code: integer, image: string} |
| 500 | 生成失败 | lxdapi_pkg_response.Response |

---

### 管理员登录

`POST` `/api/admin/login`

使用用户名、密码和验证码登录管理后台

- **认证**: 无

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| request | body | object | 是 | 登录信息 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 登录成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误或验证码错误 | lxdapi_pkg_response.Response |
| 401 | 用户名或密码错误 | lxdapi_pkg_response.Response |

---

### 管理员登出

`POST` `/api/admin/logout`

退出管理后台登录

- **认证**: SessionAuth

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 登出成功 | lxdapi_pkg_response.Response |

---

## Admin API - 容器管理

### 获取容器列表

`GET` `/api/admin/containers`

获取容器列表，可按用户筛选

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| user_id | query | string | 否 | 用户ID |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 500 | 获取失败 | lxdapi_pkg_response.Response |

---

### 获取容器详情

`GET` `/api/admin/containers/:name`

获取指定容器的详细信息

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 400 | 缺少参数 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |

---

### 删除容器

`DELETE` `/api/admin/containers/:name`

删除指定容器

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 删除任务已创建 | lxdapi_pkg_response.Response |
| 400 | 缺少参数 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |
| 500 | 删除失败 | lxdapi_pkg_response.Response |

---

### 容器操作

`POST` `/api/admin/containers/:name/action`

对容器执行操作: start/stop/restart/pause/resume/reinstall/reset-password/reset-traffic

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |
| action | query | string | 是 | 操作类型 |
| request | body | object | 否 | 操作参数 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 操作成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |
| 500 | 操作失败 | lxdapi_pkg_response.Response |

---

### 获取容器配置

`GET` `/api/admin/containers/:name/config`

获取容器的资源配置信息（仅管理员可用）

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 400 | 缺少参数 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |

---

### 更新容器配置

`PUT` `/api/admin/containers/:name/config`

热更新容器配置，支持CPU、内存、磁盘、带宽等升降级，磁盘只能扩容

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |
| request | body | models.UpdateContainerConfigRequest | 是 | 配置更新参数 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 更新成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |
| 500 | 更新失败 | lxdapi_pkg_response.Response |

---

### 获取容器凭证

`GET` `/api/admin/containers/:name/credential`

获取或创建容器访问凭证

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 400 | 缺少参数 | lxdapi_pkg_response.Response |
| 500 | 创建失败 | lxdapi_pkg_response.Response |

---

### 重新生成容器凭证

`POST` `/api/admin/containers/:name/credential`

重新生成容器访问Hash

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 生成成功 | lxdapi_pkg_response.Response |
| 400 | 缺少参数 | lxdapi_pkg_response.Response |
| 500 | 生成失败 | lxdapi_pkg_response.Response |

---

### 获取容器IP

`GET` `/api/admin/containers/:name/ip`

获取容器的IP地址

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |
| version | query | string | 否 | IP版本: v4/v6/all，默认all |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 400 | 缺少参数 | lxdapi_pkg_response.Response |
| 500 | 获取失败 | lxdapi_pkg_response.Response |

---

### 分配容器IP

`POST` `/api/admin/containers/:name/ip/allocate`

为容器分配IP地址

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |
| version | query | string | 是 | IP版本: v4 或 v6 |
| request | body | object | 否 | 分配数量 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 分配成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |
| 500 | 分配失败 | lxdapi_pkg_response.Response |

---

## Admin API - 仪表盘

### 获取仪表盘数据

`GET` `/api/admin/dashboard`

获取系统概览数据，包括容器、用户、系统信息等

- **认证**: SessionAuth

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |

---

### 获取主机状态

`GET` `/api/admin/host/stats`

获取主机CPU、内存、磁盘等实时状态

- **认证**: SessionAuth

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 500 | 获取失败 | lxdapi_pkg_response.Response |

---

## Admin API - IP管理

### 获取IP绑定列表

`GET` `/api/admin/ip`

获取IP绑定列表

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 否 | IP版本: v4/v6/all，默认all |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 500 | 获取失败 | lxdapi_pkg_response.Response |

---

### 分配IP

`POST` `/api/admin/ip/allocate`

为容器分配IP地址

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 是 | IP版本: v4 或 v6 |
| request | body | object | 是 | 分配参数 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 分配成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |
| 500 | 分配失败 | lxdapi_pkg_response.Response |

---

### 获取IP地址池

`GET` `/api/admin/ip/pool`

获取IP地址池列表

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 是 | IP版本: v4 或 v6 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |

---

### 更新IP池备注

`PUT` `/api/admin/ip/pool`

更新IP地址池中IP的备注

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 是 | IP版本: v4 或 v6 |
| request | body | object | 是 | 更新参数 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 更新成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 404 | IP不存在 | lxdapi_pkg_response.Response |
| 500 | 更新失败 | lxdapi_pkg_response.Response |

---

### 添加IP到地址池

`POST` `/api/admin/ip/pool`

添加IP地址到地址池

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 是 | IP版本: v4 或 v6 |
| request | body | object | 是 | IP信息 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 添加成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 500 | 添加失败 | lxdapi_pkg_response.Response |

---

### 从地址池删除IP

`DELETE` `/api/admin/ip/pool`

从地址池中删除IP

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 是 | IP版本: v4 或 v6 |
| request | body | object | 是 | IP地址 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 删除成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误或IP正在使用 | lxdapi_pkg_response.Response |
| 404 | IP不存在 | lxdapi_pkg_response.Response |
| 500 | 删除失败 | lxdapi_pkg_response.Response |

---

### 批量添加IP到地址池

`POST` `/api/admin/ip/pool/batch`

批量添加IP地址到地址池，v4使用start_ip+end_ip范围，v6使用start_ip+count数量

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 是 | IP版本: v4 或 v6 |
| request | body | object | 是 | IP参数 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 添加成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 500 | 添加失败 | lxdapi_pkg_response.Response |

---

### 释放IP

`POST` `/api/admin/ip/release`

释放容器的IP地址

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 是 | IP版本: v4 或 v6 |
| request | body | object | 是 | 释放参数 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 释放成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 500 | 释放失败 | lxdapi_pkg_response.Response |

---

## Admin API - NAT配置

### 获取NAT配置

`GET` `/api/admin/nat-config`

获取NAT配置

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 是 | IP版本: v4 或 v6 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |

---

### 保存NAT配置

`POST` `/api/admin/nat-config`

保存NAT配置

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 是 | IP版本: v4 或 v6 |
| request | body | object | 是 | NAT配置 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 保存成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 500 | 保存失败 | lxdapi_pkg_response.Response |

---

## Admin API - 端口映射管理

### 获取端口映射列表

`GET` `/api/admin/port-mapping`

获取端口映射列表，支持按版本和容器过滤

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 否 | IP版本: v4/v6/all，默认all |
| container | query | string | 否 | 容器名称 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 500 | 获取失败 | lxdapi_pkg_response.Response |

---

### 分配端口映射

`POST` `/api/admin/port-mapping/allocate`

为容器分配端口映射

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 是 | IP版本: v4 或 v6 |
| request | body | object | 是 | 映射参数 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 分配成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 500 | 分配失败 | lxdapi_pkg_response.Response |

---

### 释放端口映射

`POST` `/api/admin/port-mapping/release`

释放端口映射，支持单个和批量释放

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 是 | IP版本: v4 或 v6 |
| request | body | object | 是 | 释放参数 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 释放成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 500 | 释放失败 | lxdapi_pkg_response.Response |

---

## Admin API - 存储池管理

### 获取存储池列表

`GET` `/api/admin/storage-pools`

获取存储池列表

- **认证**: 无

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | OK | lxdapi_pkg_response.Response |

---

### 设置存储池优先级

`PUT` `/api/admin/storage-pools/:name/priority`

设置存储池优先级

- **认证**: 无

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 存储池名称 |
| priority | body | integer | 是 | 优先级 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | OK | lxdapi_pkg_response.Response |

---

### 从LXD同步存储池

`POST` `/api/admin/storage-pools/sync`

从LXD同步存储池

- **认证**: 无

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | OK | lxdapi_pkg_response.Response |

---

## Admin API - 任务管理

### 获取任务列表

`GET` `/api/admin/tasks`

获取任务列表，可按容器名过滤

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| container_name | query | string | 否 | 容器名称 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 500 | 获取失败 | lxdapi_pkg_response.Response |

---

### 获取任务详情

`GET` `/api/admin/tasks/:id`

获取指定任务的详细信息

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| id | path | string | 是 | 任务ID |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 400 | 缺少参数 | lxdapi_pkg_response.Response |
| 404 | 任务不存在 | lxdapi_pkg_response.Response |

---

### 删除任务

`DELETE` `/api/admin/tasks/:id`

删除指定任务，支持单个删除和批量删除

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| id | path | string | 否 | 任务ID（单个删除） |
| ids | query | string | 否 | 任务ID列表，逗号分隔（批量删除） |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 删除成功 | lxdapi_pkg_response.Response |
| 400 | 缺少参数 | lxdapi_pkg_response.Response |
| 500 | 删除失败 | lxdapi_pkg_response.Response |

---

### 批量删除任务

`POST` `/api/admin/tasks/batch-delete`

批量删除多个任务

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| request | body | object | 是 | 任务ID列表 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 删除成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 500 | 删除失败 | lxdapi_pkg_response.Response |

---

## Admin API - 模板管理

### 获取模板列表

`GET` `/api/admin/templates`

获取所有容器模板

- **认证**: SessionAuth

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 500 | 获取失败 | lxdapi_pkg_response.Response |

---

### 删除模板

`DELETE` `/api/admin/templates/:fingerprint`

删除指定模板，支持单个和批量删除

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| fingerprint | path | string | 否 | 模板指纹（单个删除） |
| fingerprints | query | string | 否 | 模板指纹列表，逗号分隔（批量删除） |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 删除成功 | lxdapi_pkg_response.Response |
| 400 | 缺少参数 | lxdapi_pkg_response.Response |
| 500 | 删除失败 | lxdapi_pkg_response.Response |

---

### 获取模板用户权限

`GET` `/api/admin/templates/:fingerprint/permission`

获取模板的用户权限设置

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| fingerprint | path | string | 是 | 模板指纹 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 404 | 模板不存在 | lxdapi_pkg_response.Response |

---

### 设置模板用户权限

`PUT` `/api/admin/templates/:fingerprint/permission`

设置哪些用户可以使用该模板

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| fingerprint | path | string | 是 | 模板指纹 |
| request | body | object | 是 | 允许使用的用户名列表 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 设置成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 404 | 模板不存在 | lxdapi_pkg_response.Response |
| 500 | 设置失败 | lxdapi_pkg_response.Response |

---

### 批量删除模板

`POST` `/api/admin/templates/batch-delete`

批量删除多个模板

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| request | body | models.BatchDeleteTemplateRequest | 是 | 模板指纹列表 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 删除成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 500 | 删除失败 | lxdapi_pkg_response.Response |

---

### 同步模板

`POST` `/api/admin/templates/sync`

从LXD同步模板到数据库

- **认证**: SessionAuth

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 同步成功 | lxdapi_pkg_response.Response |
| 500 | 同步失败 | lxdapi_pkg_response.Response |

---

## Admin API - 用户管理

### 获取用户列表或详情

`GET` `/api/admin/users`

获取用户列表，如果提供id参数则返回用户详情

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| id | query | string | 否 | 用户ID（可选，提供则返回详情） |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 404 | 用户不存在 | lxdapi_pkg_response.Response |
| 500 | 获取失败 | lxdapi_pkg_response.Response |

---

### 创建用户

`POST` `/api/admin/users`

创建新用户并生成密码

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| request | body | internal_api_admin.CreateUserRequest | 是 | 用户信息 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 创建成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 500 | 创建失败 | lxdapi_pkg_response.Response |

---

### 更新用户

`PUT` `/api/admin/users/:id`

更新用户配额和状态

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| id | path | string | 是 | 用户ID |
| request | body | internal_api_admin.UpdateUserRequest | 是 | 更新参数 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 更新成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 500 | 更新失败 | lxdapi_pkg_response.Response |

---

### 删除用户

`DELETE` `/api/admin/users/:id`

删除指定用户

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| id | path | string | 是 | 用户ID |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 删除成功 | lxdapi_pkg_response.Response |
| 400 | 缺少参数 | lxdapi_pkg_response.Response |
| 500 | 删除失败 | lxdapi_pkg_response.Response |

---

### 重新生成密码

`POST` `/api/admin/users/:id/regenerate-key`

重新生成用户的密码

- **认证**: SessionAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| id | path | string | 是 | 用户ID |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 生成成功 | lxdapi_pkg_response.Response |
| 400 | 缺少参数 | lxdapi_pkg_response.Response |
| 500 | 生成失败 | lxdapi_pkg_response.Response |

---

## Console API

### 创建WebSocket令牌

`POST` `/api/console/token`

为容器控制台创建WebSocket访问令牌

- **认证**: 无

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| request | body | object | 是 | 容器主机名 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 创建成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 500 | 生成失败 | lxdapi_pkg_response.Response |

---

### WebSocket控制台连接

`GET` `/api/console/ws`

建立WebSocket连接用于容器控制台

- **认证**: 无

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| token | query | string | 是 | 访问令牌 |

---

## Container API - 操作

### 容器操作

`POST` `/api/container/action`

对容器执行操作: start/stop/restart/reinstall/reset-password

- **认证**: ContainerAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| action | query | string | 是 | 操作类型: start/stop/restart/reinstall/reset-password |
| request | body | object | 否 | 操作参数（reinstall需要image/password，reset-password需要password） |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 操作成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |
| 500 | 操作失败 | lxdapi_pkg_response.Response |

---

## Container API - 认证

### 获取验证码

`GET` `/api/container/captcha`

获取容器登录验证码

- **认证**: 无

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 生成成功 | {captcha_id: string, code: integer, image: string} |
| 500 | 生成失败 | lxdapi_pkg_response.Response |

---

### 验证访问

`POST` `/api/container/verify`

使用Hash和验证码验证容器访问

- **认证**: 无

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| request | body | object | 是 | 验证信息 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 验证成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 401 | 验证失败 | lxdapi_pkg_response.Response |

---

## Container API - 信息

### 获取容器信息

`GET` `/api/container/info`

获取容器详细信息，包含状态和流量

- **认证**: ContainerAuth

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |

---

### 获取模板列表

`GET` `/api/container/templates`

获取可用容器模板列表

- **认证**: ContainerAuth

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 500 | 获取失败 | lxdapi_pkg_response.Response |

---

## Container API - IP管理

### 获取容器IP地址

`GET` `/api/container/ip`

获取容器的IP地址，通过version参数区分IPv4/IPv6

- **认证**: ContainerAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 否 | IP版本: v4/v6/all，默认all |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 500 | 获取失败 | lxdapi_pkg_response.Response |

---

### 分配IP地址

`POST` `/api/container/ip/allocate`

为容器分配IP地址，通过version参数区分IPv4/IPv6

- **认证**: ContainerAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 是 | IP版本: v4 或 v6 |
| request | body | object | 否 | 分配数量 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 分配成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 500 | 分配失败 | lxdapi_pkg_response.Response |

---

### 释放IP地址

`POST` `/api/container/ip/release`

释放容器的IP地址，通过version参数区分IPv4/IPv6

- **认证**: ContainerAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 是 | IP版本: v4 或 v6 |
| request | body | object | 是 | 要释放的IP地址 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 释放成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 500 | 释放失败 | lxdapi_pkg_response.Response |

---

## Container API - 端口映射

### 获取端口映射列表

`GET` `/api/container/port-mapping`

获取端口映射列表，通过version参数区分

- **认证**: ContainerAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 否 | IP版本: v4/v6/all，默认all |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 500 | 获取失败 | lxdapi_pkg_response.Response |

---

### 分配端口映射

`POST` `/api/container/port-mapping/allocate`

为容器分配端口映射，通过version参数区分IPv4/IPv6

- **认证**: ContainerAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 是 | IP版本: v4 或 v6 |
| request | body | object | 是 | 映射参数 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 分配成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 500 | 分配失败 | lxdapi_pkg_response.Response |

---

### 释放端口映射

`POST` `/api/container/port-mapping/release`

释放端口映射，通过version参数区分IPv4/IPv6

- **认证**: ContainerAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 是 | IP版本: v4 或 v6 |
| request | body | object | 是 | 映射ID列表 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 释放成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 500 | 释放失败 | lxdapi_pkg_response.Response |

---

## Public API

### 获取品牌设置

`GET` `/api/public/brand`

获取公开的品牌设置

- **认证**: 无

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |

---

## System API - 容器管理

### 获取容器列表

`GET` `/api/system/containers`

获取所有容器的列表信息

- **认证**: ApiKeyAuth

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 500 | 获取失败 | lxdapi_pkg_response.Response |

---

### 创建容器

`POST` `/api/system/containers`

创建一个新的LXD容器，支持配置CPU、内存、磁盘、带宽等资源

- **认证**: ApiKeyAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| request | body | models.CreateContainerRequest | 是 | 容器创建参数 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 创建成功，返回任务ID | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 500 | 创建失败 | lxdapi_pkg_response.Response |

---

### 获取容器详情

`GET` `/api/system/containers/{name}`

获取指定容器的详细信息和运行状态

- **认证**: ApiKeyAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 400 | 缺少容器名称 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |
| 500 | 获取失败 | lxdapi_pkg_response.Response |

---

### 删除容器

`DELETE` `/api/system/containers/{name}`

删除指定名称的容器及相关数据（危险操作）

- **认证**: ApiKeyAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 删除任务已创建 | lxdapi_pkg_response.Response |
| 400 | 缺少容器名称 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |
| 500 | 创建任务失败 | lxdapi_pkg_response.Response |

---

### 容器操作

`POST` `/api/system/containers/{name}/action`

对容器执行操作: start/stop/restart/pause/resume/reinstall/reset-password

- **认证**: ApiKeyAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |
| action | query | string | 是 | 操作类型: start/stop/restart/pause/resume/reinstall/reset-password |
| request | body | object | 否 | 操作参数（reinstall需要image/password，reset-password需要password） |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 操作成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |
| 500 | 操作失败 | lxdapi_pkg_response.Response |

---

### 更新容器配置

`PUT` `/api/system/containers/{name}/config`

热更新容器配置，支持CPU、内存、磁盘、带宽等升降级，磁盘只能扩容

- **认证**: ApiKeyAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |
| request | body | models.UpdateContainerConfigRequest | 是 | 配置更新参数 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 更新成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |
| 500 | 更新失败 | lxdapi_pkg_response.Response |

---

## System API - 凭证管理

### 获取容器访问码

`GET` `/api/system/containers/{name}/credential`

获取容器的访问码，如果不存在则自动创建

- **认证**: ApiKeyAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 400 | 缺少容器名称 | lxdapi_pkg_response.Response |
| 500 | 获取失败 | lxdapi_pkg_response.Response |

---

### 重新生成容器访问码

`POST` `/api/system/containers/{name}/credential/regenerate`

重新生成容器的访问码

- **认证**: ApiKeyAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 生成成功 | lxdapi_pkg_response.Response |
| 400 | 缺少容器名称 | lxdapi_pkg_response.Response |
| 500 | 生成失败 | lxdapi_pkg_response.Response |

---

## System API - IP管理

### 获取容器IP地址

`GET` `/api/system/ip`

获取指定容器的IP地址，通过version参数区分IPv4/IPv6

- **认证**: ApiKeyAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| container | query | string | 是 | 容器名称 |
| version | query | string | 否 | IP版本: v4/v6/all，默认all |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 400 | 缺少容器名称 | lxdapi_pkg_response.Response |
| 500 | 获取失败 | lxdapi_pkg_response.Response |

---

### 分配IP地址

`POST` `/api/system/ip/allocate`

为指定容器分配IP地址，通过version参数区分IPv4/IPv6

- **认证**: ApiKeyAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 是 | IP版本: v4 或 v6 |
| request | body | object | 是 | 分配参数 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 分配成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 500 | 分配失败 | lxdapi_pkg_response.Response |

---

### 释放IP地址

`POST` `/api/system/ip/release`

释放容器的指定IP地址回地址池，通过version参数区分IPv4/IPv6

- **认证**: ApiKeyAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 是 | IP版本: v4 或 v6 |
| request | body | object | 是 | 释放参数 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 释放成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 500 | 释放失败 | lxdapi_pkg_response.Response |

---

## System API - 端口映射管理

### 获取端口映射列表

`GET` `/api/system/port-mapping`

获取端口映射列表，通过version参数区分，支持按容器筛选

- **认证**: ApiKeyAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 否 | IP版本: v4/v6/all，默认all |
| container | query | string | 否 | 容器名称（可选） |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 500 | 获取失败 | lxdapi_pkg_response.Response |

---

### 分配端口映射

`POST` `/api/system/port-mapping/allocate`

为容器分配端口映射，通过version参数区分IPv4/IPv6

- **认证**: ApiKeyAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 是 | IP版本: v4 或 v6 |
| request | body | object | 是 | 映射参数 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 分配成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |
| 500 | 分配失败 | lxdapi_pkg_response.Response |

---

### 释放端口映射

`POST` `/api/system/port-mapping/release`

释放端口映射，通过version参数区分IPv4/IPv6

- **认证**: ApiKeyAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 是 | IP版本: v4 或 v6 |
| request | body | object | 是 | 映射ID |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 释放成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 500 | 释放失败 | lxdapi_pkg_response.Response |

---

## System API - 任务管理

### 获取任务列表

`GET` `/api/system/tasks`

获取任务列表，可按容器过滤

- **认证**: ApiKeyAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | query | string | 否 | 容器名称（可选） |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 500 | 获取失败 | lxdapi_pkg_response.Response |

---

### 获取任务详情

`GET` `/api/system/tasks/detail`

获取指定任务的详细信息

- **认证**: ApiKeyAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| id | query | string | 是 | 任务ID |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 404 | 任务不存在 | lxdapi_pkg_response.Response |

---

## System API - 流量管理

### 获取容器流量

`GET` `/api/system/traffic`

获取指定容器的流量使用情况

- **认证**: ApiKeyAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | query | string | 是 | 容器名称 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 400 | 缺少容器名称 | lxdapi_pkg_response.Response |
| 500 | 获取失败 | lxdapi_pkg_response.Response |

---

### 重置容器流量

`POST` `/api/system/traffic/reset`

重置指定容器的流量统计

- **认证**: ApiKeyAuth

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | query | string | 是 | 容器名称 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 重置成功 | lxdapi_pkg_response.Response |
| 400 | 缺少容器名称 | lxdapi_pkg_response.Response |
| 500 | 重置失败 | lxdapi_pkg_response.Response |

---

## User API - 容器管理

### 刷新缓存

`POST` `/api/user/cache/refresh`

刷新容器缓存，无name参数刷新所有，有name参数刷新指定容器

- **认证**: UserSession

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | query | string | 否 | 容器名称，不传则刷新所有 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 刷新成功 | lxdapi_pkg_response.Response |
| 403 | 无权操作 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |
| 500 | 刷新失败 | lxdapi_pkg_response.Response |

---

### 获取容器列表

`GET` `/api/user/containers`

获取当前用户的所有容器

- **认证**: UserSession

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |

---

### 创建容器

`POST` `/api/user/containers`

用户创建新容器，使用管理员预设的配置参数

- **认证**: UserSession

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| request | body | internal_api_user.UserCreateContainerRequest | 是 | 容器创建参数 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 创建成功，返回任务ID | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 500 | 创建失败 | lxdapi_pkg_response.Response |

---

### 获取容器详情

`GET` `/api/user/containers/:name`

获取指定容器的详细信息，包含状态和系统信息

- **认证**: UserSession

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 403 | 无权访问 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |

---

### 删除容器

`DELETE` `/api/user/containers/:name`

删除用户自己的容器

- **认证**: UserSession

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 删除任务已创建 | lxdapi_pkg_response.Response |
| 403 | 无权操作 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |
| 500 | 删除失败 | lxdapi_pkg_response.Response |

---

### 容器操作

`POST` `/api/user/containers/:name/action`

对容器执行操作: start/stop/restart/reinstall/reset-password

- **认证**: UserSession

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |
| action | query | string | 是 | 操作类型: start/stop/restart/reinstall/reset-password |
| request | body | object | 否 | 操作参数 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 操作成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 403 | 无权操作 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |
| 500 | 操作失败 | lxdapi_pkg_response.Response |

---

### 获取容器配置

`GET` `/api/user/containers/:name/config`

获取容器的资源配置信息

- **认证**: UserSession

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 403 | 无权访问 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |

---

### 更新容器配置

`PUT` `/api/user/containers/:name/config`

热更新容器配置，支持CPU、内存、磁盘、带宽等升降级，磁盘只能扩容

- **认证**: UserSession

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |
| request | body | models.UpdateContainerConfigRequest | 是 | 配置更新参数 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 更新成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 403 | 无权操作 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |
| 500 | 更新失败 | lxdapi_pkg_response.Response |

---

### 获取容器凭证

`GET` `/api/user/containers/:name/credential`

获取容器访问凭证，不存在则自动创建

- **认证**: UserSession

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 403 | 无权访问 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |
| 500 | 获取失败 | lxdapi_pkg_response.Response |

---

### 重新生成容器凭证

`POST` `/api/user/containers/:name/credential/regenerate`

重新生成容器访问Hash

- **认证**: UserSession

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 生成成功 | lxdapi_pkg_response.Response |
| 403 | 无权操作 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |
| 500 | 生成失败 | lxdapi_pkg_response.Response |

---

## User API - 认证

### 获取验证码

`GET` `/api/user/captcha`

获取用户登录验证码

- **认证**: 无

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 生成成功 | {captcha_id: string, code: integer, image: string} |
| 500 | 生成失败 | lxdapi_pkg_response.Response |

---

### 获取当前用户信息

`GET` `/api/user/info`

获取当前登录用户的信息和配额统计

- **认证**: UserSession

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |

---

### 用户登录

`POST` `/api/user/login`

使用用户名、密码和验证码登录

- **认证**: 无

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| request | body | object | 是 | 登录信息 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 登录成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 401 | 认证失败 | lxdapi_pkg_response.Response |
| 500 | 登录失败 | lxdapi_pkg_response.Response |

---

### 用户登出

`POST` `/api/user/logout`

退出用户中心登录

- **认证**: UserSession

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 登出成功 | lxdapi_pkg_response.Response |

---

## User API - IP管理

### 获取容器IP

`GET` `/api/user/containers/:name/ip`

获取容器的IP地址，通过version参数区分IPv4/IPv6

- **认证**: UserSession

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |
| version | query | string | 否 | IP版本: v4/v6/all，默认all |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 403 | 无权访问 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |
| 500 | 获取失败 | lxdapi_pkg_response.Response |

---

### 分配容器IP

`POST` `/api/user/containers/:name/ip/allocate`

为容器分配IP地址，通过version参数区分IPv4/IPv6

- **认证**: UserSession

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |
| version | query | string | 是 | IP版本: v4 或 v6 |
| request | body | object | 否 | 分配数量 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 分配成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 403 | 无权操作 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |
| 500 | 分配失败 | lxdapi_pkg_response.Response |

---

### 释放容器IP

`POST` `/api/user/containers/:name/ip/release`

释放容器的IP地址，通过version参数区分IPv4/IPv6

- **认证**: UserSession

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| name | path | string | 是 | 容器名称 |
| version | query | string | 是 | IP版本: v4 或 v6 |
| request | body | object | 是 | 要释放的IP地址 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 释放成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 403 | 无权操作 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |
| 500 | 释放失败 | lxdapi_pkg_response.Response |

---

## User API - 端口映射

### 获取端口映射列表

`GET` `/api/user/port-mapping`

获取端口映射列表，通过version参数区分IPv4/IPv6

- **认证**: UserSession

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 否 | IP版本: v4/v6/all，默认all |
| container_name | query | string | 否 | 容器名称，不传则返回用户所有端口映射 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 403 | 无权访问 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |
| 500 | 获取失败 | lxdapi_pkg_response.Response |

---

### 分配端口映射

`POST` `/api/user/port-mapping/allocate`

为容器分配端口映射，通过version参数区分IPv4/IPv6

- **认证**: UserSession

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 是 | IP版本: v4 或 v6 |
| request | body | object | 是 | 端口映射参数 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 分配成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 403 | 无权操作 | lxdapi_pkg_response.Response |
| 404 | 容器不存在 | lxdapi_pkg_response.Response |
| 500 | 分配失败 | lxdapi_pkg_response.Response |

---

### 释放端口映射

`POST` `/api/user/port-mapping/release`

释放端口映射，通过version参数区分IPv4/IPv6

- **认证**: UserSession

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| version | query | string | 是 | IP版本: v4 或 v6 |
| request | body | object | 是 | 端口映射ID列表 |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 释放成功 | lxdapi_pkg_response.Response |
| 400 | 参数错误 | lxdapi_pkg_response.Response |
| 403 | 无权操作 | lxdapi_pkg_response.Response |
| 404 | 端口映射不存在 | lxdapi_pkg_response.Response |
| 500 | 释放失败 | lxdapi_pkg_response.Response |

---

## User API - 任务管理

### 获取任务状态

`GET` `/api/user/tasks/:id`

获取指定任务的状态（只能查询自己的任务）

- **认证**: UserSession

**参数**

| 参数名 | 位置 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- | --- |
| id | path | string | 是 | 任务ID |

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 403 | 无权访问 | lxdapi_pkg_response.Response |
| 404 | 任务不存在 | lxdapi_pkg_response.Response |

---

## User API - 模板管理

### 获取模板列表

`GET` `/api/user/templates`

获取当前用户可用的容器模板列表

- **认证**: UserSession

**响应**

| 状态码 | 描述 | Schema |
| --- | --- | --- |
| 200 | 获取成功 | lxdapi_pkg_response.Response |
| 500 | 获取失败 | lxdapi_pkg_response.Response |

---

## 数据模型 (Definitions)

### `internal_api_admin.CreateUserRequest`

```json
{
  "type": "object",
  "required": [
    "username"
  ],
  "properties": {
    "allow_nesting": {
      "type": "boolean"
    },
    "cpu_allowance": {
      "type": "integer"
    },
    "cpu_quota": {
      "type": "integer"
    },
    "disk_quota": {
      "type": "integer"
    },
    "egress": {
      "type": "integer"
    },
    "ingress": {
      "type": "integer"
    },
    "io_read": {
      "type": "integer"
    },
    "io_write": {
      "type": "integer"
    },
    "ipv4_mapping_limit": {
      "type": "integer"
    },
    "ipv4_pool_limit": {
      "type": "integer"
    },
    "ipv6_mapping_limit": {
      "type": "integer"
    },
    "ipv6_pool_limit": {
      "type": "integer"
    },
    "max_cpu_per_container": {
      "type": "integer"
    },
    "memory_quota": {
      "type": "integer"
    },
    "memory_swap": {
      "type": "boolean"
    },
    "processes_limit": {
      "type": "integer"
    },
    "reverse_proxy_limit": {
      "type": "integer"
    },
    "traffic_limit": {
      "type": "integer"
    },
    "username": {
      "type": "string"
    }
  }
}
```

### `internal_api_admin.UpdateUserRequest`

```json
{
  "type": "object",
  "properties": {
    "allow_nesting": {
      "type": "boolean"
    },
    "cpu_allowance": {
      "type": "integer"
    },
    "cpu_quota": {
      "type": "integer"
    },
    "disk_quota": {
      "type": "integer"
    },
    "egress": {
      "type": "integer"
    },
    "ingress": {
      "type": "integer"
    },
    "io_read": {
      "type": "integer"
    },
    "io_write": {
      "type": "integer"
    },
    "ipv4_mapping_limit": {
      "type": "integer"
    },
    "ipv4_pool_limit": {
      "type": "integer"
    },
    "ipv6_mapping_limit": {
      "type": "integer"
    },
    "ipv6_pool_limit": {
      "type": "integer"
    },
    "max_cpu_per_container": {
      "type": "integer"
    },
    "memory_quota": {
      "type": "integer"
    },
    "memory_swap": {
      "type": "boolean"
    },
    "processes_limit": {
      "type": "integer"
    },
    "reverse_proxy_limit": {
      "type": "integer"
    },
    "status": {
      "type": "string"
    },
    "traffic_limit": {
      "type": "integer"
    }
  }
}
```

### `internal_api_user.UserCreateContainerRequest`

```json
{
  "type": "object",
  "required": [
    "image",
    "name"
  ],
  "properties": {
    "cpu": {
      "type": "integer"
    },
    "disk": {
      "type": "integer"
    },
    "image": {
      "type": "string"
    },
    "ipv4_mapping_limit": {
      "type": "integer"
    },
    "ipv4_pool_limit": {
      "type": "integer"
    },
    "ipv6_mapping_limit": {
      "type": "integer"
    },
    "ipv6_pool_limit": {
      "type": "integer"
    },
    "memory": {
      "type": "integer"
    },
    "name": {
      "type": "string"
    },
    "password": {
      "type": "string"
    },
    "reverse_proxy_limit": {
      "type": "integer"
    },
    "traffic_limit": {
      "type": "integer"
    }
  }
}
```

### `lxdapi_pkg_response.Response`

```json
{
  "type": "object",
  "properties": {
    "code": {
      "type": "integer"
    },
    "data": {},
    "msg": {
      "type": "string"
    }
  }
}
```

### `models.BatchDeleteTemplateRequest`

```json
{
  "type": "object",
  "required": [
    "fingerprints"
  ],
  "properties": {
    "fingerprints": {
      "type": "array",
      "items": {
        "type": "string"
      }
    }
  }
}
```

### `models.BrandSettings`

```json
{
  "type": "object",
  "properties": {
    "admin_bg_image": {
      "type": "string"
    },
    "admin_bg_opacity": {
      "type": "integer"
    },
    "admin_content_opacity": {
      "type": "integer"
    },
    "admin_login_title": {
      "type": "string"
    },
    "admin_system_name": {
      "type": "string"
    },
    "admin_system_title": {
      "type": "string"
    },
    "container_base_template": {
      "type": "string"
    },
    "container_bg_image": {
      "type": "string"
    },
    "container_bg_opacity": {
      "type": "integer"
    },
    "container_content_opacity": {
      "type": "integer"
    },
    "container_lite_template": {
      "type": "string"
    },
    "container_login_title": {
      "type": "string"
    },
    "container_notice": {
      "type": "string"
    },
    "container_notice_opacity": {
      "type": "integer"
    },
    "container_system_name": {
      "type": "string"
    },
    "container_system_title": {
      "type": "string"
    },
    "created_at": {
      "type": "string"
    },
    "favicon_url": {
      "type": "string"
    },
    "footer_text": {
      "type": "string"
    },
    "id": {
      "type": "integer"
    },
    "tls_cert_content": {
      "type": "string"
    },
    "tls_key_content": {
      "type": "string"
    },
    "updated_at": {
      "type": "string"
    },
    "use_local_cdn": {
      "type": "boolean"
    },
    "user_bg_image": {
      "type": "string"
    },
    "user_bg_opacity": {
      "type": "integer"
    },
    "user_content_opacity": {
      "type": "integer"
    },
    "user_login_title": {
      "type": "string"
    },
    "user_notice": {
      "type": "string"
    },
    "user_notice_opacity": {
      "type": "integer"
    },
    "user_system_name": {
      "type": "string"
    },
    "user_system_title": {
      "type": "string"
    }
  }
}
```

### `models.CreateContainerRequest`

```json
{
  "type": "object",
  "required": [
    "image",
    "name"
  ],
  "properties": {
    "allow_nesting": {
      "type": "boolean"
    },
    "cpu": {
      "type": "integer"
    },
    "cpu_allowance": {
      "type": "integer"
    },
    "disk": {
      "type": "integer"
    },
    "egress": {
      "type": "integer"
    },
    "image": {
      "type": "string"
    },
    "ingress": {
      "type": "integer"
    },
    "io_read": {
      "type": "integer"
    },
    "io_write": {
      "type": "integer"
    },
    "ipv4_mapping_limit": {
      "type": "integer"
    },
    "ipv4_pool_limit": {
      "type": "integer"
    },
    "ipv6_mapping_limit": {
      "type": "integer"
    },
    "ipv6_pool_limit": {
      "type": "integer"
    },
    "memory": {
      "type": "integer"
    },
    "memory_swap": {
      "type": "boolean"
    },
    "name": {
      "type": "string"
    },
    "password": {
      "type": "string"
    },
    "privileged": {
      "type": "boolean"
    },
    "processes_limit": {
      "type": "integer"
    },
    "remark": {
      "type": "string"
    },
    "reverse_proxy_limit": {
      "type": "integer"
    },
    "traffic_limit": {
      "type": "integer"
    },
    "username": {
      "type": "string"
    }
  }
}
```

### `models.UpdateContainerConfigRequest`

```json
{
  "type": "object",
  "properties": {
    "allow_nesting": {
      "type": "boolean"
    },
    "cpu": {
      "type": "integer"
    },
    "cpu_allowance": {
      "type": "integer"
    },
    "disk": {
      "type": "integer"
    },
    "egress": {
      "type": "integer"
    },
    "ingress": {
      "type": "integer"
    },
    "io_read": {
      "type": "integer"
    },
    "io_write": {
      "type": "integer"
    },
    "ipv4_mapping_limit": {
      "type": "integer"
    },
    "ipv4_pool_limit": {
      "type": "integer"
    },
    "ipv6_mapping_limit": {
      "type": "integer"
    },
    "ipv6_pool_limit": {
      "type": "integer"
    },
    "memory": {
      "type": "integer"
    },
    "memory_swap": {
      "type": "boolean"
    },
    "privileged": {
      "type": "boolean"
    },
    "processes_limit": {
      "type": "integer"
    },
    "remark": {
      "type": "string"
    },
    "reverse_proxy_limit": {
      "type": "integer"
    },
    "traffic_limit": {
      "type": "integer"
    }
  }
}
```
