# buck-ccm

buck-ccm 是一个降压变换器（buck converter）CCM/DCM 平均模型核算工具。用户给出输入电压 Vin、占空比 D、电感 L、电容 C、开关周期 Ts 和负载电阻 R（SI 单位），它按理想开关伏秒平衡判定工作模式（CCM/DCM），给出输出电压 Vout、无量纲电感系数 K、临界 Kcrit、电感电流纹波 Δi_L、电容电压纹波 Δv_C，以及一个开关周期内的电感电流三角波点列。它只做电力电子稳态核算，不做 EMI/热/效率这类附加模型，也不做升降压（buck-boost）或磁路（Φ=NI/Σℜ）核算。

## 用法

一条可复现命令（预置算例 `example/12v-5v.json`：12V 输入、占空比 5/12，应判 CCM，Vout≈5V）：

```text
go run . mode example/12v-5v.json
```

输出模式、Vout、K、Kcrit、纹波与边界量。其它子命令：

```text
buck-ccm ripple example/12v-5v.json              电感电流三角波点列与纹波
buck-ccm design example/12v-5v.json --vout 5.0   反求达到目标 Vout 的占空比并回读
buck-ccm check example/12v-5v.json               交叉规则自查（详见下方）
buck-ccm -http :8080                             启动 HTTP 服务（web 前端 + /api/*）
buck-ccm version / help
```

### HTTP 与前端

`go run . -http :8080` 后浏览器打开 <http://localhost:8080>。页面可「加载示例」填入 `example/12v-5v.json`，点「计算」会分别请求后端：

- `POST /api/mode`（JSON 参数）→ `{"mode":"CCM","vout":5.0,"k":2,"kcrit":0.5833,...}`
- `POST /api/ripple` → 纹波量与 `points` 电感电流三角波点列（页面直接绘制这些点）

参数非法时返回 400 + error JSON，例如 `{"error":"参数 d 非法：占空比必须在开区间 (0,1) 内（得到 1.5）","field":"d"}`，前端会原样显示。

### 算例格式

`example/12v-5v.json` 结构：

```json
{
  "name": "12v-5v",
  "vin": 12, "d": 0.4167, "l": 0.0001,
  "c": 0.00022, "ts": 0.00001, "r": 10
}
```

`name`/`note` 是可选的标注字段，不参与核算。六个物理参数缺一不可。

## 关键约定

- **模式判定**：K = 2L/(R·Ts)，Kcrit(D) = 1−D。K > Kcrit 判 CCM（Vout = D·Vin）；否则判 DCM。恰好落在边界（K = Kcrit）时两种模型预测重合。
- **DCM 输出**：DCM 的电压比 M 是方程 K·M² + D²·M − D² = 0 在 (D,1) 的根，恒有 Vout > D·Vin（低于同 D 的 CCM 预测才违反物理）。求解器用二分法，默认容差 1e-10、最多 120 次迭代，不收敛会报错。
- **纹波**：Δi_L = (Vin − Vout)·D·Ts/L，必须减 Vout；电容电压纹波用 Δi 对 C 积分（ESR 钉为 0）。
- **参数边界**：D 必须在开区间 (0,1)，D=0 或 D=1 视为非法；Vin、L、C、Ts、R 必须为正且有限。任何非法输入一律报错，绝不静默给数。
- **交叉规则自查**（`check`）：L 减到边界以下 → 翻 DCM 且 Vout 高于 D·Vin；D 加倍（仍 CCM 且未过压）→ Vout 加倍；Ts 加倍（仍 CCM）→ Δi_L 加倍；边界 K=Kcrit 处两模式一致。

## 构建与测试

```text
go build ./...
go test ./...
go vet ./...
```

## 目录

- `internal/spec/` — 输入参数模型、校验与 JSON 读写
- `internal/engine/` — 模式判定、稳态求解、纹波、波形点列、交叉规则、扫描与反向设计
- `internal/server/` — 薄 HTTP 层（/api/* 与静态资源）
- `example/` — 离线小算例
- `web/` — 由 Go 同进程提供的前端页面
