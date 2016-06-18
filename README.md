#rcmp
rcmp是一个对实验报告进行比较和输出的工具，以GPLv3发布。

# 编译
rcmp只提供golang代码，不提供binary（就是这么傲娇）。

## Linux
- 首先，在 [golang.org](https://golang.org) 下载go安装包。
- 设置`GOPATH`环境变量到一个目录，比如`~/golang`。
- 执行`go get github.com/femrat/rcmp`获取源码。
- 执行`go build github.com/femrat/rcmp`以编译。
- 在当前目录下即可得到`rcmp`。

## Windows
等我过些日子装上 Windows 再说…

## Others
请自求多福，能编译就应该能用。

# 使用
## 推荐场景及命名
假设有一个数据集 dataset，其中有很多个实例文件（instance-file）。
当评价一个算法及参数组合是否足够好时，可以使用这个算法和这套参数，分别输入 dataset 中所有的 instance-file。
每个 instance-file 被输入后，都会由算法输出一些关于解的信息。
一个常见的解格式为：`instance-file opt time`，其中 instance-file 是被输入的实例，opt 是解质量，time 是求解时间。

一个示例report文件是这样的（数据来源：[MaxSAT Evaluation 2015](http://www.maxsat.udl.cat/15/detailed/incomplete-ms-industrial-table.html)）：

	divider-problem.dimacs_11.filtered.cnf	2	32.42
	divider-problem.dimacs_2.filtered.cnf	2	14.12
	divider-problem.dimacs_5.filtered.cnf	2	55.19
	divider-problem.dimacs_8.filtered.cnf	2	41.29
	mem_ctrl-problem.dimacs_27.filtered.cnf	-1	-1


## 输入的 report 文件格式
report 文件的文件名以参数方式传入。
report 文件应包括至少一行。
其中每行都应可以使用 TAB 分割成多列（TAB 是默认分割符，可以修改）。

其中，每行的第一列必须为 instance-file，可以是独立的文件名，也可以带有目录。
对于 instance-file 之外的列，需要满足所选 engine 的要求。不同 engine 对列要求也不同。


## 数据流

reports --> 【预处理器 preprocessor】 --> 【比对引擎 engine】 --> 【模板 template】 --> output

通过命令行输入的 reports，经过预处理器 preprocessor 进行过滤、格式化，然后输入比对引擎 engine 进行数据处理，
处理出来的原始数据交给模板 template 渲染，然后输出到`stdout`(`pipe 1`)。
模板全权负责输出样式，engine 只提供数据，完全不干涉输出。

## 预处理器 preprocessor

**未完待续**


