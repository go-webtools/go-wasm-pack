package main

import (
    "fmt"
    "io"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
)

func main() {
    // 创建 build 目录
    buildDir := "build"
    os.MkdirAll(buildDir, os.ModePerm)

    // 查找包含 main 函数的 Go 文件
    goFile, err := findMainGoFile()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error finding main Go file: %s\n", err)
        os.Exit(1)
    }

    // 设置输出文件名
    outputWasm := filepath.Join(buildDir, "output.wasm")

    // 构建命令
    cmd := exec.Command("go", "build", "-o", outputWasm)

    // 设置环境变量
    cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")
    cmd.Args = append(cmd.Args, goFile)

    // 执行命令并获取输出
    outputBytes, err := cmd.CombinedOutput()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %s\n", err)
        os.Exit(1)
    }

    // 输出构建过程信息
    fmt.Println(string(outputBytes))
    fmt.Printf("Compiled %s to %s\n", goFile, outputWasm)

    // 复制 wasm_exec.js
    if err := copyWasmExec(buildDir); err != nil {
        fmt.Fprintf(os.Stderr, "Failed to copy wasm_exec.js: %s\n", err)
        os.Exit(1)
    }

    fmt.Printf("Copied wasm_exec.js to %s\n", filepath.Join(buildDir, "wasm_exec.js"))
}

// findMainGoFile 查找当前目录下的 Go 文件中是否有 main 函数
func findMainGoFile() (string, error) {
    files, err := os.ReadDir(".")
    if err != nil {
        return "", err
    }

    for _, file := range files {
        if strings.HasSuffix(file.Name(), ".go") {
            content, err := os.ReadFile(file.Name())
            if err == nil && strings.Contains(string(content), "func main()") {
                return file.Name(), nil
            }
        }
    }

    return "", fmt.Errorf("no main Go file found")
}

// copyWasmExec 复制 wasm_exec.js 到指定目录
func copyWasmExec(destDir string) error {
    originGoPath := os.Getenv("goPath")
    goPath := strings.Replace(originGoPath, "path", "", 1)
    if goPath == "" {
        goPath = filepath.Join(os.Getenv("HOME"), "go") // 默认 goPath
    }
    wasmExecPath := filepath.Join(goPath, "misc", "wasm", "wasm_exec.js")

    if _, err := os.Stat(wasmExecPath); os.IsNotExist(err) {
        return fmt.Errorf("wasm_exec.js not found at %s\n", wasmExecPath)
    }

    destWasmExec := filepath.Join(destDir, "wasm_exec.js")
    sourceFile, err := os.Open(wasmExecPath)
    if err != nil {
        return err
    }
    defer sourceFile.Close()

    destinationFile, err := os.Create(destWasmExec)
    if err != nil {
        return err
    }
    defer destinationFile.Close()

    _, err = io.Copy(destinationFile, sourceFile)
    return err
}
