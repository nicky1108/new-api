# openhubs 用户使用指南

本文面向普通用户，说明如何注册登录、创建 API Key、调用模型、查看用量、充值、购买订阅以及管理账号安全。管理员入口和系统配置不在本文范围内。

如果你的页面布局与本文截图或入口名称略有不同，通常是因为站点管理员切换了前端主题、隐藏了部分模块，或关闭了在线充值、订阅、异步任务等功能。本文以当前默认前端为主；classic 旧主题的对应路径通常以 `/console/` 开头。

## 快速上手

首次使用建议按以下顺序完成：

1. 打开站点并注册账号：`/sign-up`
2. 登录账号：`/sign-in`
3. 进入「个人资料」检查邮箱、安全设置和可用模型：`/profile`
4. 进入「API Keys」创建 API Key：`/keys`
5. 进入「Playground」测试模型是否可用：`/playground`
6. 在自己的应用或代码中配置 `base_url` 和 API Key
7. 进入「Usage Logs」查看调用记录和扣费明细：`/usage-logs/common`
8. 余额不足时进入「Wallet」充值或兑换：`/wallet`

## 账号注册与登录

### 注册账号

1. 访问 `/sign-up`。
2. 填写用户名、密码和邮箱。
3. 如果页面要求邮箱验证码，点击「发送验证码」，到邮箱中复制验证码并填回页面。
4. 提交注册表单，注册成功后即可进入控制台。

如果站点开启了邀请机制，注册时可能需要填写邀请码。邀请码一般由已有用户或管理员提供。

### 登录账号

1. 访问 `/sign-in`。
2. 使用用户名和密码登录。
3. 如果站点开启了第三方登录，可选择页面中的 GitHub、Discord、LinuxDO 等 OAuth 登录入口。
4. 如果已经开启 2FA，登录时还需要输入验证器 App 中的动态验证码。

### 忘记密码

在登录页点击「忘记密码」，输入绑定邮箱。系统会发送重置链接或验证码，按邮件提示完成新密码设置。重置成功后，旧密码会失效。

## 个人资料与账号安全

进入 `/profile` 管理账号资料、安全设置和通知。

### 基本资料

你可以在个人资料页修改用户名、绑定邮箱或更改密码。修改密码时需要输入当前密码、新密码和确认密码。

### 双因素认证

建议为重要账号开启 2FA：

1. 安装 Google Authenticator、Microsoft Authenticator 或其他兼容 TOTP 的验证器 App。
2. 在个人资料页找到双因素认证区域，点击开启。
3. 使用验证器扫描二维码。
4. 输入 App 中生成的 6 位验证码完成绑定。
5. 妥善保存备用码。备用码通常只展示一次，用于手机丢失或无法使用验证器时恢复登录。

### Passkey 无密码登录

如果浏览器和设备支持 Passkey，可以在个人资料页注册 Passkey。注册后可使用指纹、面容识别、系统密码或硬件安全密钥登录。

### 第三方账号绑定

在第三方账号区域绑定 GitHub、Discord 等账号后，下次可使用对应平台快速登录。解绑前请确认仍有其他可用登录方式。

### 可用模型

个人资料页通常会展示当前账号可调用的模型列表。复制模型名称时请保留完整字符串，例如 `gpt-5.4-mini` 或管理员配置的自定义模型名。

## 创建和管理 API Key

API Key 是调用 openhubs API 的凭证。进入 `/keys` 创建和管理 API Key。

### 创建 API Key

1. 点击「Create API Key」。
2. 填写名称，建议按用途命名，例如 `local-dev`、`production-web`、`cursor`。
3. 按需设置配额、过期时间、模型限制和 IP 白名单。
4. 提交后立即复制完整 Key。

完整 Key 通常只在创建时展示一次。关闭弹窗后，如果页面只显示脱敏内容，需要重新复制或重新创建。

### 常用限制项

| 配置项 | 用途 | 建议 |
| --- | --- | --- |
| 配额 | 限制单个 Key 最多可消耗的额度 | 给测试、外部工具、团队成员单独设置上限 |
| 过期时间 | 到期后 Key 自动失效 | 临时用途设置短期限，长期服务定期轮换 |
| 模型限制 | 限制 Key 只能调用指定模型 | 防止误用高成本模型 |
| IP 白名单 | 限制允许请求的来源 IP | 生产服务可配合固定出口 IP 使用 |

IP 白名单依赖网关、反向代理和真实客户端 IP 配置。不要只依赖它防护泄露的 Key。

### 复制连接信息

API Key 列表中的行操作通常提供：

- 复制 Key
- 复制连接信息
- 导入到聊天应用
- 编辑限制
- 删除 Key

如果怀疑 Key 泄露，请立即删除或停用该 Key，并在相关应用中替换为新 Key。

## 使用 Playground 测试模型

进入 `/playground` 可直接测试模型，无需写代码。

1. 选择一个可用模型。
2. 在输入框中输入消息。
3. 点击发送，等待模型回复。
4. 如果请求失败，检查模型是否可用、API Key 是否有效、账户余额是否充足。

Playground 适合在接入代码前验证账号、模型、渠道和计费是否正常。

## 调用 API

openhubs 提供兼容 OpenAI 风格的接口。接入时通常只需要替换两项：

- `base_url`：改为 openhubs 的 API 地址，通常为 `https://openhubs.xyz/v1`
- `api_key`：填写你在 `/keys` 创建的 API Key

### Base URL 写法

不同工具对 Base URL 的字段名略有差异，但填写内容相同：

| 场景 | 填写方式 |
| --- | --- |
| OpenAI SDK | `base_url="https://openhubs.xyz/v1"` |
| 聊天客户端 | `API Base URL` 填 `https://openhubs.xyz/v1` |
| 直接 HTTP 调用 | 请求完整地址以 `https://openhubs.xyz/v1/...` 开头 |

不要把接口路径重复写两次。例如 SDK 里已经配置了 `https://openhubs.xyz/v1`，调用聊天接口时 SDK 会自动请求 `/chat/completions`，不需要把 Base URL 写成 `https://openhubs.xyz/v1/chat/completions`。

### API Key 鉴权

所有 API 请求都需要在 HTTP Header 中携带 Bearer Token：

```http
Authorization: Bearer sk-your-openhubs-key
Content-Type: application/json
```

其中 `sk-your-openhubs-key` 替换为你在 `/keys` 创建的完整 API Key。不要把 Key 放在 URL 查询参数中，也不要写进前端公开代码或 Git 仓库。

可以先用模型列表接口验证 Base URL 和 Key 是否正确：

```bash
curl https://openhubs.xyz/v1/models \
  -H "Authorization: Bearer sk-your-openhubs-key"
```

如果返回可用模型列表，说明地址和鉴权基本正常。

### Python：单轮对话

```python
from openai import OpenAI

client = OpenAI(
    api_key="sk-your-openhubs-key",
    base_url="https://openhubs.xyz/v1",
)

response = client.chat.completions.create(
    model="gpt-5.4-mini",
    messages=[
        {"role": "user", "content": "你好，请用一句话介绍 openhubs。"}
    ],
)

print(response.choices[0].message.content)
```

### cURL：单轮对话

```bash
curl https://openhubs.xyz/v1/chat/completions \
  -H "Authorization: Bearer sk-your-openhubs-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-5.4-mini",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ]
  }'
```

### cURL：多轮对话

多轮对话需要把历史消息一起发送给接口。常见顺序是 `system`、`user`、`assistant`、`user`：

```bash
curl https://openhubs.xyz/v1/chat/completions \
  -H "Authorization: Bearer sk-your-openhubs-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-5.4-mini",
    "messages": [
      {"role": "system", "content": "你是一个简洁、准确的中文助手。"},
      {"role": "user", "content": "openhubs 的 Base URL 是什么？"},
      {"role": "assistant", "content": "Base URL 是 https://openhubs.xyz/v1。"},
      {"role": "user", "content": "那 API Key 应该放在哪里？"}
    ]
  }'
```

### cURL：流式对话

如果客户端支持 Server-Sent Events，可以设置 `stream: true` 获取边生成边返回的结果：

```bash
curl -N https://openhubs.xyz/v1/chat/completions \
  -H "Authorization: Bearer sk-your-openhubs-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-5.4-mini",
    "stream": true,
    "messages": [
      {"role": "user", "content": "用三点说明如何保护 API Key。"}
    ]
  }'
```

### Python：流式对话

```python
from openai import OpenAI

client = OpenAI(
    api_key="sk-your-openhubs-key",
    base_url="https://openhubs.xyz/v1",
)

stream = client.chat.completions.create(
    model="gpt-5.4-mini",
    stream=True,
    messages=[
        {"role": "user", "content": "写一段 100 字以内的产品介绍。"}
    ],
)

for chunk in stream:
    delta = chunk.choices[0].delta.content
    if delta:
        print(delta, end="")
```

### cURL：图片生成

图片生成使用 `POST /v1/images/generations`。模型名以你的可用模型或定价页展示为准；如果管理员配置的是其他图像模型，请替换示例中的 `gpt-image-2`。

```bash
curl https://openhubs.xyz/v1/images/generations \
  -H "Authorization: Bearer sk-your-openhubs-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-image-2",
    "prompt": "一张干净现代的 openhubs 控制台宣传图，浅色背景，科技感",
    "size": "1024x1024",
    "n": 1
  }'
```

接口通常会返回图片 URL 或 Base64 图片数据，具体取决于上游模型和管理员配置。请根据返回结果中的 `url` 或 `b64_json` 字段保存图片。

### Python：图片生成并保存 Base64 图片

```python
import base64
from openai import OpenAI

client = OpenAI(
    api_key="sk-your-openhubs-key",
    base_url="https://openhubs.xyz/v1",
)

image = client.images.generate(
    model="gpt-image-2",
    prompt="一枚 openhubs 风格的应用图标，简洁，白底，绿色点缀",
    size="1024x1024",
    n=1,
)

item = image.data[0]
if getattr(item, "b64_json", None):
    with open("openhubs-image.png", "wb") as f:
        f.write(base64.b64decode(item.b64_json))
else:
    print(item.url)
```

### 常用接口

| 能力 | 路径 | 用途 |
| --- | --- | --- |
| 聊天补全 | `POST /v1/chat/completions` | 多轮对话、问答、内容生成 |
| 文本补全 | `POST /v1/completions` | 旧式补全接口 |
| Embeddings | `POST /v1/embeddings` | 文本向量化 |
| 图像生成 | `POST /v1/images/generations` | 文生图 |
| 语音识别 | `POST /v1/audio/transcriptions` | 音频转文字 |
| 语音合成 | `POST /v1/audio/speech` | 文字转语音 |
| 模型列表 | `GET /v1/models` | 查询当前 Key 可用模型 |

不同模型支持的参数可能不同。生产接入前请先在 Playground 或测试环境中验证模型、上下文长度、流式输出和错误处理。

## 接入聊天应用

很多聊天客户端支持自定义 OpenAI 兼容接口。通用配置如下：

| 配置项 | 填写内容 |
| --- | --- |
| API Base URL | `https://openhubs.xyz/v1` |
| API Key | 在 `/keys` 创建的 Key |
| Model | 在个人资料、定价页或模型列表中复制的模型名 |

### 一键导入

在 `/keys` 的行操作中，如果看到「导入」或「连接信息」按钮，可选择支持的聊天应用。系统会生成跳转链接或配置内容。

### 手动配置

如果目标应用不支持一键导入：

1. 打开目标应用设置。
2. 找到「OpenAI Compatible」「自定义 API」「API Provider」等配置区域。
3. 填写 API Base URL、API Key 和模型名。
4. 保存后发送一条测试消息。

常见可接入应用包括 Lobe Chat、OpenCat、AI as Workspace 以及其他支持 OpenAI 兼容接口的客户端。

## 查看使用记录

进入 `/usage-logs/common` 查看普通 API 调用记录。

日志通常包含：

- 调用时间
- 模型名称
- API Key 或令牌名称
- 输入与输出 Token
- 消耗配额
- 请求状态
- 错误信息

可以按时间范围、模型、Key、请求状态等条件筛选。排查问题时，先复制请求时间、模型、错误信息，再联系管理员会更高效。

### 绘图和异步任务记录

如果站点启用了绘图或异步任务功能，可在使用记录中切换到：

- `/usage-logs/drawing`
- `/usage-logs/task`

任务状态常见含义：

| 状态 | 含义 |
| --- | --- |
| `PENDING` | 已提交，等待处理 |
| `IN_PROGRESS` | 正在生成 |
| `SUCCESS` | 已完成，可查看结果 |
| `FAILURE` | 生成失败，通常会按规则退还或修正配额 |

## 钱包、充值与兑换码

进入 `/wallet` 查看余额、充值、兑换码、邀请奖励和订阅计划。

### 在线充值

如果管理员开启了在线支付：

1. 在钱包页选择充值金额，或输入自定义金额。
2. 选择支付方式，例如 EPay、Stripe、Creem、Waffo 等。
3. 确认金额并跳转支付。
4. 支付完成后返回 openhubs，等待余额刷新。

如果支付成功但余额未更新，请保留支付订单号和付款截图，联系管理员核对。

### 兑换码充值

1. 在钱包页找到兑换码输入框。
2. 粘贴管理员提供的兑换码。
3. 点击兑换。
4. 兑换成功后余额会增加。

兑换码可能有有效期、额度和使用次数限制。提示失效时请联系发码方确认。

### 邀请返利

钱包页可能显示专属邀请链接。你可以复制链接邀请新用户注册。被邀请用户产生消费后，返利额度会进入奖励余额；按页面提示转入主余额后即可使用。

## 订阅计划

如果站点配置了订阅计划，钱包页会展示可购买套餐和当前订阅状态。

订阅适合稳定用量场景。购买前请确认：

- 套餐价格
- 有效期
- 包含配额或权益
- 配额重置周期
- 是否允许重复购买
- 是否支持自动续费

如果同时拥有钱包余额和订阅额度，页面可能提供扣费偏好设置，例如优先使用订阅额度或优先使用钱包余额。实际扣费规则以页面展示和管理员配置为准。

## 查看模型定价

进入 `/pricing` 查看模型价格。

你可以：

1. 搜索模型名称。
2. 查看输入价格、输出价格和特殊计费项。
3. 对比不同模型的成本。
4. 根据实际需求选择低成本或高能力模型。

实际扣费可能受到用户分组倍率、模型倍率、订阅抵扣、缓存 Token、图片/音频计费等因素影响。最终以使用记录中的扣费明细为准。

## 常见问题

### 为什么提示余额不足？

可能原因：

- 钱包余额不足。
- API Key 设置了单独配额上限。
- 订阅额度已用完。
- 模型价格较高，预扣额度超过当前可用余额。

处理方式：到 `/wallet` 充值或兑换，到 `/keys` 检查 Key 配额限制，到 `/pricing` 查看模型价格。

### 为什么某个模型不可用？

可能原因：

- 当前账号分组没有该模型权限。
- API Key 设置了模型限制。
- 上游渠道暂时不可用。
- 管理员禁用了该模型。

处理方式：在 `/profile` 查看可用模型，在 `/keys` 检查模型限制，并使用 `/playground` 测试。

### 为什么请求被限流？

可能是站点全局限流、用户分组限流或 API Key 级限流生效。降低请求频率、增加重试退避，或联系管理员调整限流策略。

### API Key 泄露了怎么办？

立即进入 `/keys` 删除或停用泄露的 Key，然后创建新 Key 并更新到应用配置中。不要把 Key 提交到 Git 仓库、公开文档、前端代码或截图中。

### 如何排查 API 调用失败？

按顺序检查：

1. API Base URL 是否为 `https://openhubs.xyz/v1`
2. API Key 是否复制完整
3. 模型名是否正确
4. 余额是否充足
5. Key 是否过期、被限额或限制模型
6. `/usage-logs/common` 中是否有错误详情

## classic 旧主题路径对照

如果站点使用 classic 主题，常见入口可能是：

| 功能 | default 路径 | classic 路径 |
| --- | --- | --- |
| 控制台首页 | `/dashboard` | `/console` |
| API Key | `/keys` | `/console/token` |
| Playground | `/playground` | `/console/playground` |
| 钱包 | `/wallet` | `/console/topup` |
| 个人设置 | `/profile` | `/console/personal` |
| 使用记录 | `/usage-logs/common` | `/console/log` |
| 任务记录 | `/usage-logs/task` | `/console/task` |
| 定价 | `/pricing` | `/pricing` |

## 参考来源

本文结合本项目当前路由和以下 New API 用户指南整理：

- https://docs.newapi.pro/zh/docs/guide/feature-guide
- https://docs.newapi.pro/zh/docs/guide/feature-guide/user/auth
- https://docs.newapi.pro/zh/docs/guide/feature-guide/user/personal-setting
- https://docs.newapi.pro/zh/docs/guide/feature-guide/user/token
- https://docs.newapi.pro/zh/docs/guide/feature-guide/user/api
- https://docs.newapi.pro/zh/docs/guide/feature-guide/user/chat-apps
- https://docs.newapi.pro/zh/docs/guide/feature-guide/user/log
- https://docs.newapi.pro/zh/docs/guide/feature-guide/user/topup
- https://docs.newapi.pro/zh/docs/guide/feature-guide/user/subscription
- https://docs.newapi.pro/zh/docs/guide/feature-guide/user/pricing
- https://docs.newapi.pro/zh/docs/guide/feature-guide/user/task
