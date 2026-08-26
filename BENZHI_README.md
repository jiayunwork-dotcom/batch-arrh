# batch-arrh: Go headless HTTP API and CLI for isothermal batch reactor conversion

batch-arrh integrates first-order, second-order and A→B→C series rate laws (direct k or Arrhenius A/Ea/T) for a constant-volume batch, and serves the same kernels over HTTP.

读入等温恒容批式反应器 JSON 场景（反应级数、速率常数 k 或 Arrhenius 指前因子 A/活化能 Ea/温度 T、初始浓度与停留时间），核算转化率 X、各物种浓度与摩尔数。给定 k 时直接积分 ODE，给定 A/Ea/T 时先按 k=A·exp(−Ea/(R·T)) 求速率再积分；一阶与等摩尔二阶场景下用 Damköhler 数闭式解与积分结果交叉校验，串联 A→B→C 与不等摩尔二阶另有解析恒等式。非法浓度、非正速率或越界转化率在求解前拒绝；结果可写入版本化 JSON 快照，空文件或截断 JSON 读回时报错。

## How to run

```
go run .
go run . serve
go run . integrate example/first-order.json
```

## API

- `GET /api/health` — liveness
- `POST /api/integrate` — scenario JSON → conversion, concentrations, k
- `POST /api/verify` — same JSON → closed-form vs integrator check

Listen on `:8080`. Illegal domain values return HTTP 422.

## 评测镜像

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh batch-arrh
```
