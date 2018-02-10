# MatrixFS

本项目是在我们[基于多斜率码链阵列纠删码](http://www.joca.cn/CN/abstract/abstract20400.shtml)的[验证Demo](https://github.com/Vaaaas/Array-fault-tolerant-system)的基础上，使用Golang 重新编写的多节点版本。

本项目不再使用本地单个计算机中的不同文件夹作为不同的存储节点，而是使用不同的计算机作为不同的存储节点，在同一局域网中实现分布式存储及灾后恢复的功能。
同时，从之前Demo 的Windows Form 应用程序的形式改为了由命令行启动中心节点和存储节点，在Web 端进行系统管理和对文件的存取等操作。

## 特点

* 理论上不受限制的容错能力
* 运算效率上较RS码的提升超过2个数量级
* 效率能随着条块尺寸的增加而提高（固定容错能力）

## 使用说明

