# 轻轻量级日志监控程序

## 功能
监控指定日志文件，发现指定关键词，通过钉钉推送报警（钉钉机器人）

## 使用
- 项目根目录运行 `sh makefile.sh` 会生成`bin`目录
- 将`bin`文件夹中的内容拷贝到服务器中即可
- `./bin/conf/`是配置文件目录，具体配置 参考示例文件
- 后台运行logNotice文件即可： `./logNotice &` 