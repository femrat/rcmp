#rcmp
rcmp 是一个对实验报告进行比较和输出的工具，以 GPLv3 发布。

# 编译
rcmp 只提供 golang 代码，不提供 binary（就是这么傲娇）。

## Linux
- 首先，在 [golang.org](https://golang.org) 下载 go 安装包。
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

预处理器按照先后顺序，分为1) dup-check, 2) basename, 3) suffix, 4) filter, 5) filter-sort, 6) intersection。
其中，在第二步和第三步之间会做一个安全检查，确保每个 report 中都没有重复的 instance-file。

- 1) dup-check：
检查输入的所有 report 文件是否有重复，如果发现重复，仅保留第一个出现的 report。
例如，`rcmp ... a b a b c a`与`rcmp ... a b c`效果一致。
该功能默认开启，可以指定`-dup-report`参数关闭。

- 2）basename：
输入的 report 中的所有 instance-file，如果包含目录，目录都会被剥离，只保留文件名。
该功能默认开启，可以指定`-keep-basename`参数关闭。

- 3) suffix：
如果指定了`-trim-suffix SUFFIX`参数，那么输入的 report 中的所有 instance-file，如果以`SUFFIX`结尾，SUFFIX都会被剥离。

- 4) filter：
如果指定了`-filter-file FILTER-FILE`参数，那么`FILTER-FILE`中的每一行都会被认为是一个允许出现的 instance-file。
在每个 report 中，未在`FILTER-FILE`中出现的 instance-file 将被移除。
需要注意的是，在 filter 过程中，`FILTER-FILE`中存在不在任何 report 中出现的 instance-file 是允许的，filter 后的 report 可以含有比`FILTER-FILE`更少的 instance-file。

- 5) filter-sort：
该参数必须与`-filter-file`同时使用。
filter-sort 会将每个 report 中的 instance-file 按照`FILTER-FILE`中出现的顺序重新排序。
如果`FILTER-FILE`中存在 report 中缺少的 instance-file，则该 instance-file 不会影响排序，filter-sort 会自动忽略它，排序完成后 report 中也不会存在空洞。
每个 report 都独立进行排序，排序后改变的仅有顺序。

- 6) intersection:
如指定`-intersection`参数，intersection操作会在所有 report 中求交集，将交集之外的 instance-file 全部移除。
该操作不会对 instance-file 的出现顺序做出改变，仅保证所有 report 拥有相同的 instance-file。

## 比对引擎 engine

engine 运行之前，会做一个安全检查：检查所有 report 是否具有相同的 instance-file 列表。
检查时，顺序不一致也会报错。

检查通过后，根据选择的 engine，执行对应步骤，并将 engine 的输出传给 template 进行渲染输出。

engine 中可能出现以下选项：

- template 选择：`-t TEMPLATE`。其中`TEMPLATE`是模板的文件路径。
rcmp 会尝试打开`TEMPLATE`，如果失败，则会尝试在 rcmp 所在目录下的 template/ 目录中寻找，如果仍未找到这个文件，会报错。

- 比对模式：`-mode MODE`。目前该选项仅有一个有效输入且为默认，所以无需指定。
`MODE`目前仅支持`sat`，并且其也是该选项的默认值。指定其他值均将报错。
比对模式决定了如何界定解的优劣和有效性。


### engine: s (Compare separately)
这个 engine 将输入的第一个 report 作为基础，将其他 report 依次独立地与其相比。
除了`-t`和`-mode`外，该 engine 还支持替换 instance-file。
如果给出`-rename FILE`参数，那么`FILE`文件中的替换规则将被采用，输出给模板的 instance-file 将被重命名。
其中，`FILE`中的每一行应遵守如下规则：`old-instance-file new-instance-file`。
如：

	divider-problem.dimacs_11.filtered.cnf	divider-problem_11
	divider-problem.dimacs_2.filtered.cnf	divider-problem_2
	divider-problem.dimacs_5.filtered.cnf	divider-problem_5
	divider-problem.dimacs_8.filtered.cnf	divider-problem_8
	mem_ctrl-problem.dimacs_27.filtered.cnf	mem_ctrl-problem_27

需要注意，这里的`old-instance-file`是预处理器处理后的名字，不是原始 report 文件中的名字。
如果预处理器中做了`-trim-suffix`等操作，那么此处`old-instance-name`应是处理后的名字。
`FILE`中行的顺序没有影响，但是不允许相同的`old-instance-file`出现两次。

支持的模板：`s`，`s.csv`，`s.tex`，分别为控制台样式，csv 和 latex。

### engine: sg (Compare separately and group by given rules.)
这个 engine 与 s 类似，都是将输入的第一个 report 作为基础，将其他 report 依次独立地与其相比。
只是该 engine 会将比对的结果按照指定的分组规则合并。

sg 的必选参数是`-group GROUP-FILE`。其中`GROUP-FILE`是分组规则，格式为`group-name instance-file`，如：

	divider  divider-problem.dimacs_11.filtered.cnf
	divider  divider-problem.dimacs_2.filtered.cnf
	divider  divider-problem.dimacs_5.filtered.cnf
	divider  divider-problem.dimacs_8.filtered.cnf
	mem_ctrl mem_ctrl-problem.dimacs_27.filtered.cnf

这里严格要求`GROUP-FILE`中的每个 instanec-file 都能和 report 中的对应。
如果预处理器中做了`-trim-suffix`等操作，那么此处`instance-name`应是处理后的名字。
template 收到的分组顺序，以`GROUP-FILE`中的每个分组第一次出现的顺序为准。

支持的模板：`sg`，`sg.csv`，`sg.tex`，分别为控制台样式，csv 和 latex。

