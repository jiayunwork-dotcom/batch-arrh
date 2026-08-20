# batch-arrh

batch-arrh 是等温间歇反应器的转化率核算命令行工具：给定初始浓度、计量系数、速率常数（直接 k 或阿伦尼乌斯参数 A/Ea/T，k=A·exp(−Ea/(R·T))）与停留时间，对声明为等容液相的反应器做物料衡算，输出关键物种转化率 X、各物种浓度与摩尔数，以及连串反应（A→B→C）中间产物的选择率与峰时刻。

- 输入：JSON 场景文件（`initial_concentration`、`stoichiometry`、`rate`、`volume`、`residence_time`），示例见 `example/`；命令行 `--t` 可覆盖停留时间，`--steps` 可覆盖积分步数
- 输出：`integrate` 子命令打印转化率、浓度与摩尔数；`--table` 打印转化率-时间轨迹表；`--json` 输出机器可读 JSON（含完整轨迹）
- 速率：一级 −r_A=k·C_A 用解析解 X=1−exp(−k·t)；二级 −r_A=k·C_A·C_B 用 RK4 对转化率方程积分（等初始浓度时闭式 1/C_A−1/C_A0=k·t 可核对）；连串 A→B→C 用解析解并钉死摩尔守恒 C_A+C_B+C_C=C_A0
- 边界：等容液相（C_A=C_A0·(1−X)，无变容项）；物料衡算 (−r_A)·V=dN_A/dt 与速率定律共用同一计量；不涉及变容气相、并联多步网络外的全厂流程、闪蒸或精馏
- 非法参数（k<0、T≤0、Ea<0、初始浓度 <0、t<0、k 与阿伦尼乌斯同时给出、缺少速率参数）一律以 error 返回，CLI 打印 stderr 并非零退出；t=0 合法并返回初值

## 钉死的约定

- **等容关系**：`C_A = C_A0·(1−X)`，物种换算统一为 `C_i = C_i0 − (ν_i/ν_A)·C_A0·X`，反应物下降、产物上升，任何代码路径不引入变容因子。
- **共用计量**：速率定律只消费浓度，`dX/dt = (−r_A)/C_A0` 由 (−r_A)·V = −dN_A/dt 在 V 恒定下推出，V 仅在摩尔数核算（N=C·V）中出现。
- **同一温度与 R**：阿伦尼乌斯求 k、速率求值、积分共用同一 `thermo.R`（8.314 J/(mol·K)）与开尔文温标；Ea 单位固定为 J/mol。
- **一级闭式**：X=1−exp(−k·t) 既用于积分也用于交叉核对；`--t 0` 时该式与轨迹都回到初值 X=0。
- **连串守恒**：C_C 恒由 `C_A0−C_A−C_B` 推出，故 A+B+C 的摩尔守恒在任何时刻都精确成立；B 峰时刻解析为 t_max=ln(k1/k2)/(k1−k2)（k1≠k2）。
- **校验**：`k<0`、`T≤0`、`Ea<0`、初始浓度 <0、`t<0`、速率参数缺失或 `k` 与 `arrhenius` 同时给出，都在 `internal/reactor` 的 New 阶段以 error 返回。

## 可测契约

- 一级：`t` 加倍时 `1−X` 变成原来的平方；升温（同 A、Ea）使 k 与同样 t 的 X 上升。
- 二级等初始浓度：`1/C_A − 1/C_A0 = k·t` 在最终状态成立；积分轨迹每一点都满足等容关系 `C_A = C_A0·(1−X)`（`internal/reactor` 的 `TestSecondOrderMatchesInverseConcentration`、`TestSecondOrderConstantVolumeRelation`）。
- 连串：`C_A+C_B+C_C=C_A0` 每时刻成立；轨迹峰值时刻与解析 t_max 差不超过一个积分步长（`TestSeriesConservesMoles`、`TestSeriesPeakTimeMatchesAnalytic`）。

## 构建 / 运行 / 测试

```text
go build ./...                     # 编译（纯标准库）
go test ./...                      # 全部测试（thermo / kinetics / reactor / report）
go run . integrate example/first-order.json --t 100
go run . integrate example/second-order.json --t 100
go run . integrate example/series.json --table
go run . integrate example/first-order-arrhenius.json --json
```

一级算例核对：k=0.05，t=100，X=1−exp(−5)=0.993262，C_A=2·exp(−5)=0.0134759 mol/L。
