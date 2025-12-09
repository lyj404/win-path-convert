package main

import (
	"fmt"
	"os"

	"github.com/lyj404/win-path-convert/internal/app"
)

func main() {
	// 调用应用程序的启动函数
	if err := app.RunApplication(); err != nil {
		// 如果启动或运行过程中发生错误，输出错误信息
		fmt.Printf("应用程序错误: %v\n", err)
		os.Exit(1)
	}
}
