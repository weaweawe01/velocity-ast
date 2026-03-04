package main

import (
	"fmt"
	"os"
	"time"

	"github.com/weaweawe01/velocity-ast/internal/dump"
	"github.com/weaweawe01/velocity-ast/internal/parser"
)

func main() {
	// 开始毫秒级时间
	start := time.Now()
	tpl := "#set($e=666);$e.getClass().forName(\"java.lang.Runtime\").getMethod(\"getRuntime\",null).invoke(null,null).exec(\"calc\")"
	root, tokens, err := parser.Parse(tpl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse error: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(dump.Render(root, tokens))
	elapsed := time.Since(start).Milliseconds()
	fmt.Printf("耗时：%d 毫秒\n", elapsed)
}
