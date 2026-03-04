# velocity-ast

一个用 Go 语言实现的 [Apache Velocity](https://velocity.apache.org/) 模板语言 **AST（抽象语法树）解析器**，其输出与 Java 官方实现严格兼容。

---

## 简介

`github.com/weaweawe01/velocity-ast` 将 Velocity 模板字符串解析为 AST，并以与 Java 参考实现（`velocity-java`）完全一致的格式输出树形转储结果。这使其可以作为安全分析、静态检测或模板预处理场景下的高性能替代方案。

### 核心特性

- **Go 原生实现**：无 JVM 依赖，启动快、资源占用低
- **Java 兼容输出**：AST 节点 ID、Token 范围、树形前缀符号与 Java 基线严格一致
- **完整指令覆盖**：支持 `#set`、`#if / #elseif / #else`、`#foreach`、`#macro`、`#define`、`#parse`、`#include`、`#stop`、`#break`、`#evaluate` 等所有核心指令
- **回归测试套件**：内置多组测试用例，可与 Java 输出一键对比

---


## 快速开始
```shell
go get github.com/weaweawe01/velocity-ast
```

## 解析模板并输出 AST
```
package main
import (
	"fmt"
	"os"
	"time"

	velocity "github.com/weaweawe01/velocity-ast"
)

func main() {
	// 开始毫秒级时间
	start := time.Now()
	tpl := "#set($e=666);$e.getClass().forName(\"java.lang.Runtime\").getMethod(\"getRuntime\",null).invoke(null,null).exec(\"calc\")"
	root, tokens, err := velocity.Parse(tpl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(velocity.Render(root, tokens))
	elapsed := time.Since(start).Milliseconds()
	fmt.Printf("耗时：%d 毫秒\n", elapsed)
}

```

## 执行效果
```shell
root@hcss-ecs-5ed3:~/goggg# go run . 
ASTprocess [id=0, info=0, invalid=false, tokens=[#set(], [$e], [=], [666], [)], [;], [$e], [.], [g...] -> #set(
├── ASTSetDirective [id=26, info=0, invalid=false, tokens=[#set(], [$e], [=], [666], [)]] -> #set(
│   ├── ASTReference [id=20, info=0, invalid=false, tokens=[$e]] -> $e
│   └── ASTExpression [id=27, info=0, invalid=false, tokens=[666]] -> 666
│       └── ASTIntegerLiteral [id=8, info=0, invalid=false, tokens=[666]] -> 666
├── ASTText [id=2, info=0, invalid=false, tokens=[;]] -> ;
└── ASTReference [id=20, info=0, invalid=false, tokens=[$e], [.], [getClass], [(], [)], [.], [forName], [...] -> $e
    ├── ASTMethod [id=18, info=0, invalid=false, tokens=[getClass], [(], [)]] -> getClass
    │   └── ASTIdentifier [id=10, info=0, invalid=false, tokens=[getClass]] -> getClass
    ├── ASTMethod [id=18, info=0, invalid=false, tokens=[forName], [(], ["java.lang.Runtime"], [)]] -> forName
    │   ├── ASTIdentifier [id=10, info=0, invalid=false, tokens=[forName]] -> forName
    │   └── ASTExpression [id=27, info=0, invalid=false, tokens=["java.lang.Runtime"]] -> "java.lang.Runtime"
    │       └── ASTStringLiteral [id=9, info=0, invalid=false, tokens=["java.lang.Runtime"]] -> "java.lang.Runtime"
    ├── ASTMethod [id=18, info=0, invalid=false, tokens=[getMethod], [(], ["getRuntime"], [,], [null], [)]] -> getMethod
    │   ├── ASTIdentifier [id=10, info=0, invalid=false, tokens=[getMethod]] -> getMethod
    │   ├── ASTExpression [id=27, info=0, invalid=false, tokens=["getRuntime"]] -> "getRuntime"
    │   │   └── ASTStringLiteral [id=9, info=0, invalid=false, tokens=["getRuntime"]] -> "getRuntime"
    │   └── ASTExpression [id=27, info=0, invalid=false, tokens=[null]] -> null
    │       └── ASTReference [id=20, info=0, invalid=false, tokens=[null]] -> null
    ├── ASTMethod [id=18, info=0, invalid=false, tokens=[invoke], [(], [null], [,], [null], [)]] -> invoke
    │   ├── ASTIdentifier [id=10, info=0, invalid=false, tokens=[invoke]] -> invoke
    │   ├── ASTExpression [id=27, info=0, invalid=false, tokens=[null]] -> null
    │   │   └── ASTReference [id=20, info=0, invalid=false, tokens=[null]] -> null
    │   └── ASTExpression [id=27, info=0, invalid=false, tokens=[null]] -> null
    │       └── ASTReference [id=20, info=0, invalid=false, tokens=[null]] -> null
    └── ASTMethod [id=18, info=0, invalid=false, tokens=[exec], [(], ["calc"], [)]] -> exec
        ├── ASTIdentifier [id=10, info=0, invalid=false, tokens=[exec]] -> exec
        └── ASTExpression [id=27, info=0, invalid=false, tokens=["calc"]] -> "calc"
            └── ASTStringLiteral [id=9, info=0, invalid=false, tokens=["calc"]] -> "calc"

```



## 项目结构

```
github.com/weaweawe01/velocity-ast/
├── cmd/
│   └── github.com/weaweawe01/velocity-ast-astdump/   # CLI 工具：输出模板 AST
├── docs/
│   └── ast-output-spec.md     # AST 输出格式规范
├── internal/
│   ├── ast/                   # AST 节点类型定义
│   ├── lexer/                 # 词法分析器
│   ├── parser/                # 语法解析器
│   └── dump/                  # AST 树形渲染器
├── scripts/                   # 对比 / 回归测试脚本
├── testdata/
│   ├── cases/                 # .vtl 测试用例
│   └── expected-java/         # Java 基线输出
├── tests/
│   └── parser_test.go         # Go 测试入口
└── velocity-java/             # Java 参考实现（用于基线生成）
```

---



## velocity-java — AST 基准生成工具

`velocity-java/` 是一个 Maven 子项目，基于 **Apache Velocity Engine 2.3** 官方 Java 实现，用于生成供 Go 实现对比的 AST 基线输出。

### 目录结构

```
| 程序 | 用途 |
|---|---|
| `VelocityAstDump` | **基线来源**，输出格式与 Go 实现严格比对，是回归测试的权威参考 |
| `VelocityAstDumpPretty` | 缩进风格输出，便于人工阅读 AST 结构 |
| `VelocityAstTreeDemo` | 简单树形展示，用于快速调试 |

```

## 许可证

本项目遵循 [Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0)。
