# buck-ccm：Go 编写的降压变换器 CCM/DCM 平均模型核算工具，带 Web 与 HTTP API

buck-ccm 读入输入电压 Vin、占空比 D、电感 L、电容 C、开关周期 Ts 与负载电阻 R（SI 单位），判定 CCM/DCM 模式并给出 Vout、K、Kcrit、电感电流纹波与电容电压纹波。

## 构建 / 运行 / 测试

```text
go build ./...
go run . mode example/12v-5v.json
go test ./...
```

其他子命令与 HTTP 启动方式见项目 `README.md`：`ripple`、`design`、`check`、`-http :8080`、`version`、`help`。

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
