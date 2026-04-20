# package_demo

用于演示 `ialang build` 与 `ialang run-pkg`：

```bash
# 1) 直接运行源码
go run ./cmd/ialang run examples/package_demo/main.ia

# 2) 编译并打包（会把依赖模块一起打进去）
go run ./cmd/ialang build examples/package_demo/main.ia -o examples/package_demo/app.iapkg

# 3) 直接执行字节码包
go run ./cmd/ialang run-pkg examples/package_demo/app.iapkg

# 4) 打包为独立二进制（通过在编译器尾部追加包数据）
go run ./cmd/ialang build-bin examples/package_demo/main.ia -o examples/package_demo/package_demo.exe
./examples/package_demo/package_demo.exe
```

预期输出（关键行）：

```text
== package-demo v1.0.0 ==
sync: sum=14, product=45
async: sum=7, product=12
```
