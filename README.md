# mto
MTO is a powerful information gathering tool specifically designed for cybersecurity professionals. It integrates modules from Hunter, Fofa, and Quake, making it convenient for users to extract valuable information from various data sources.
MTO-Tool 是一个强大的资产信息提取工具，支持从 FOFA、Hunter 和 Quake 等平台提取数据。该工具通过命令行界面提供使用便利性，适用于安全研究人员和网络管理员
## 使用方法

### 基本使用

### 初次运行配置

程序初次运行时，会在用户主目录下创建 `mto/config` 文件夹，需要配置 API 以正常使用工具。请按照提示进行配置。

启动 MTO-Tool：

```sh
mto.exe
```

这将显示工具的主界面和可用命令列表。

### 可用命令

- `hunter`: MTO 的 Hunter 模块，用于从 Hunter 提取资产信息。
- `fofa`: MTO 的 FOFA 提取模块，用于从 FOFA 提取资产信息。
- `quake`: MTO 的 Quake 提取模块，用于从 Quake 提取资产信息。
- `help`: 显示关于任何命令的帮助信息。

### FOFA 模块示例

从 FOFA 提取资产信息：

```sh
mto.exe fofa [flags]
```

#### 可用 Flags

- `-s, --search string`: 单个 FOFA 语法查询。
- `-f, --file string`: 从本地文件读取 FOFA 语法。
- `-o, --output string`: 输出 `-f` 参数结果到 CSV 文件，默认输出到 `fofa.csv`。
- `-u, --url`: 过滤输出 URL 信息。
- `-ip`: 过滤输出 IP 信息。
- `-d,--max int`: 最大结果数量，默认为1000，单次查询最大支持获取10000条结果（仅在普通查询时有效）。
- `-n,--next`: 使用连续翻页专业接口，避免数据错位问题，会自动获取所有可用结果，需要额外消耗F币（不受 -d 参数限制）。
- `-k, --k`: 查询 FOFA 语法。
- `-h, --help`: 显示帮助信息。

#### 使用示例

1. **使用单个 FOFA 语法查询**：

   ```sh
   mto.exe fofa -s 'title="登录" && country="CN"'
   mto.exe fofa -s title="登录"
   ```

2. **从本地文件读取 FOFA 语法**：

   ```sh
   mto.exe fofa -f fofa_queries.txt
   ```

3. **输出结果到指定的 CSV 文件**：

   ```sh
   mto.exe fofa -f fofa_queries.txt -o custom_output.csv
   ```

4. **过滤输出 URL 信息跟其他扫描工具配合使用**：

   ```sh
   mto.exe fofa -s title="登录" -u | nuclei.exe
   ```

5. **过滤输出 IP 信息**：

   ```sh
   mto.exe fofa -s title="登录" -ip
   ```

6. **查询 FOFA 语法**：

   ```sh
   mto.exe fofa -k
   ```

7. **指定最大结果数量**：

   ```sh
   mto.exe fofa -s title="登录" -d 5000  # 获取5000条结果（最大查询大小是10000条）
   mto.exe fofa -s title="登录"           # 默认获取1000条结果（去重前）
   ```

8. **使用连续翻页接口**：

   ```sh
   mto.exe fofa -s title="登录" -n    # 使用连续翻页接口获取所有结果，如数据量超过10000条使用这个参数，单次查询结果超过10,000条，超出部分可能需要额外消耗F币（不受 -m 参数限制）
   ```

   > 注意：使用 `-n` 参数时，`-d` 参数将被忽略，系统会自动获取所有可用结果。

