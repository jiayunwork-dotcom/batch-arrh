# batch-arrh

batch-arrh 是等温间歇反应器的转化率核算命令行工具：给定初始浓度、计量系数、速率常数（直接 k 或阿伦尼乌斯参数 A/Ea/T，k=A·exp(−Ea/(R·T))）与停留时间，对声明为等容液相的反应器做物料衡算，输出关键物种转化率 X、各物种浓度与摩尔数，以及连串反应（A→B→C）中间产物的选择率与峰时刻。

## 构建 / 运行 / 测试

```text
go build ./...     # 编译
go run . integrate example/first-order.json --t 100
go test ./...      # 测试
```

## 评测镜像

本目录评测专用文件（勿覆盖项目自带 Dockerfile/README）：

- `benzhi.Dockerfile`
- `build_benzhi_docker.sh`
- `BENZHI_README.md`（本文件）

两种架构都要构建并进容器验证：

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh <image-name> linux/arm64
./build_benzhi_docker.sh <image-name> linux/amd64
docker run -it <image-name>:latest
```
