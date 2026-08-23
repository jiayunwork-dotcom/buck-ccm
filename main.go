// Command buck-ccm 是降压变换器（buck）CCM/DCM 平均模型核算工具。
//
// 用户以 JSON 给出输入电压 Vin、占空比 D、电感 L、电容 C、开关周期 Ts
// 与负载电阻 R，工具判定导通模式（CCM/DCM），给出 Vout、无量纲电感
// 系数 K、临界 Kcrit、电感电流纹波、电容电压纹波与电感电流三角波点列，
// 并可按交叉规则自查。
//
// 子命令：
//
//	buck-ccm mode <算例.json>     判定模式并输出稳态（Vout/K/Kcrit）
//	buck-ccm ripple <算例.json>   输出纹波与电感电流三角波点列
//	buck-ccm design <算例.json> --vout 5   反求占空比并回读校验
//	buck-ccm check <算例.json>    交叉规则自查（L 减半 / D 加倍 / Ts 加倍 / 边界）
//	buck-ccm -http :8080          以 HTTP 服务启动（web 前端 + /api/*）
//	buck-ccm version              打印版本
//	buck-ccm help                 显示帮助
//
// 所有错误写入 stderr 并以非零退出码结束；非法输入绝不静默给出数值。
// 求解逻辑全部在 internal/，本文件只做参数接线。
package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"buck-ccm/internal/engine"
	"buck-ccm/internal/server"
	"buck-ccm/internal/spec"
)

// version 是 CLI 版本号。
const version = "1.0.0"

// usageText 是帮助文本。
const usageText = `buck-ccm —— 降压变换器 CCM/DCM 平均模型核算

用法:
  buck-ccm mode <算例.json>       判定模式并输出稳态（Vout/K/Kcrit）
  buck-ccm ripple <算例.json>     输出纹波与电感电流三角波点列
  buck-ccm design <算例.json> --vout 5   反求占空比并回读校验
  buck-ccm check <算例.json>      交叉规则自查
  buck-ccm -http :8080            以 HTTP 服务启动（web 前端 + /api/*）
  buck-ccm version                打印版本
  buck-ccm help                   显示本帮助

算例示例:
  buck-ccm mode example/12v-5v.json
  buck-ccm -http :8080

算例 JSON 字段（SI 单位）:
  vin  输入电压（V，必须为正）
  d    占空比（开区间 (0,1)）
  l    电感（H，必须为正）
  c    电容（F，必须为正）
  ts   开关周期（s，必须为正）
  r    负载电阻（Ω，必须为正）

说明:
  K = 2L/(R·Ts)，Kcrit(D) = 1−D。K > Kcrit 判 CCM（Vout = D·Vin），
  否则 DCM（Vout 高于同 D 的 CCM 预测）。
  CCM 纹波 Δi_L = (Vin−Vout)·D·Ts/L。非法参数一律报错并以非零退出码结束。
`

func main() {
	httpAddr := flag.String("http", "", "以 HTTP 服务启动并监听该地址（如 :8080）")
	flag.Usage = func() { fmt.Fprint(os.Stderr, usageText) }
	flag.Parse()

	if *httpAddr != "" {
		startHTTP(*httpAddr)
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		fail("缺少子命令，运行 buck-ccm help 查看用法")
	}
	switch args[0] {
	case "mode":
		requireArg(args, "mode")
		runMode(args[1])
	case "ripple":
		requireArg(args, "ripple")
		runRipple(args[1])
	case "design":
		runDesign(args[1:])
	case "check":
		requireArg(args, "check")
		runCheck(args[1])
	case "version", "-v", "--version":
		fmt.Printf("buck-ccm %s\n", version)
	case "help", "-h", "--help":
		fmt.Print(usageText)
	default:
		fail("未知子命令 %q，运行 buck-ccm help 查看用法", args[0])
	}
}

// requireArg 检查子命令是否带了算例路径参数。
func requireArg(args []string, cmd string) {
	if len(args) < 2 {
		fail("%s 需要一个算例文件参数，如 example/12v-5v.json", cmd)
	}
}

// startHTTP 启动 HTTP 服务：校验静态目录后监听地址。
func startHTTP(addr string) {
	if err := server.EnsureStaticDirs(server.DefaultWebDir, server.DefaultExampleDir); err != nil {
		fail("%v", err)
	}
	fmt.Printf("buck-ccm HTTP 服务监听 %s（前端 web/，API /api/mode /api/ripple）\n", addr)
	handler := server.New()
	if err := http.ListenAndServe(addr, handler); err != nil {
		fail("HTTP 服务启动失败：%v", err)
	}
}

// runMode 加载算例、核算模式与稳态并打印报告。
func runMode(path string) {
	s, err := spec.LoadFile(path)
	if err != nil {
		fail("%v", err)
	}
	res, err := engine.Analyze(*s)
	if err != nil {
		fail("%v", err)
	}
	fmt.Print(res.Report())
}

// runRipple 加载算例、核算纹波并打印纹波量与前几个波形点。
func runRipple(path string) {
	s, err := spec.LoadFile(path)
	if err != nil {
		fail("%v", err)
	}
	res, err := engine.Analyze(*s)
	if err != nil {
		fail("%v", err)
	}
	wave, err := engine.InductorCurrentWaveformDefault(*s, res.Mode, res.Vout)
	if err != nil {
		fail("%v", err)
	}
	fmt.Printf("== 纹波核算 ==\n")
	fmt.Printf("模式: %s\n", res.Mode)
	fmt.Printf("Δi_L = %s，Δv_C = %s\n", spec.FormatSI(res.DeltaIL, "A"), spec.FormatSI(res.DeltaVC, "V"))
	fmt.Printf("平均负载电流 = %s，电感电流峰值 = %s，续流 D2 = %s\n",
		spec.FormatSI(res.Iavg, "A"), spec.FormatSI(res.Ipeak, "A"), spec.FormatSI(res.D2, ""))
	fmt.Printf("电感电流三角波点列（%d 点，周期 %s）:\n", len(wave.Points), spec.FormatSI(wave.Period, "s"))
	for _, p := range wave.Points {
		fmt.Printf("  t=%-14s i=%s\n", spec.FormatSI(p.T, "s"), spec.FormatSI(p.I, "A"))
	}
}

// runDesign 解析 --vout 与算例路径，输出反向设计（反求占空比）并回读校验。
func runDesign(args []string) {
	path := ""
	var voutTarget float64
	hasTarget := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--vout":
			if i+1 >= len(args) {
				fail("design --vout 缺少数值")
			}
			if _, err := fmt.Sscanf(args[i+1], "%g", &voutTarget); err != nil {
				fail("design --vout 需要数值，实际 %q", args[i+1])
			}
			hasTarget = true
			i++
		default:
			if path != "" {
				fail("design 多余参数 %q", args[i])
			}
			path = args[i]
		}
	}
	if path == "" {
		fail("design 需要一个算例文件参数，如 example/12v-5v.json")
	}
	if !hasTarget {
		fail("design 需要 --vout <目标电压>")
	}
	s, err := spec.LoadFile(path)
	if err != nil {
		fail("%v", err)
	}
	d, err := engine.DesignDuty(*s, voutTarget)
	if err != nil {
		fail("%v", err)
	}
	check, err := engine.VerifyDesign(*s, voutTarget)
	if err != nil {
		fail("%v", err)
	}
	fmt.Printf("== 反向设计：%s ==\n", path)
	fmt.Printf("目标 Vout = %s，设计占空比 D = %s\n",
		spec.FormatSI(voutTarget, "V"), spec.FormatSI(d, ""))
	fmt.Printf("回读：D=%.6g 下模式 %s，Vout=%s，相对偏差 %.3g\n",
		check.D, check.Mode, spec.FormatSI(check.Vout, "V"), check.Deviation)
	fmt.Printf("当前算例：K=%.4g，Kcrit=%.4g，临界电感 Lcrit=%s\n",
		engine.ParameterK(*s), engine.Kcrit(*s),
		spec.FormatSI(engine.CriticalInductance(*s), "H"))
}

// runCheck 加载算例并执行交叉规则自查，任何规则 FAIL 都以非零退出码结束。
func runCheck(path string) {
	s, err := spec.LoadFile(path)
	if err != nil {
		fail("%v", err)
	}
	results, err := engine.RunChecks(*s)
	if err != nil {
		fail("%v", err)
	}
	fmt.Print(engine.FormatChecks(results))
	if engine.CheckFailed(results) {
		os.Exit(1)
	}
}

// fail 把错误写入 stderr 并以退出码 1 结束。
func fail(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "buck-ccm: "+format+"\n", a...)
	os.Exit(1)
}
