Go 实现简单的分布式系统
# 服务启动
先启动registryService

再启动loggerService

接着启动gradeService

最后启动portal

# Web端

浏览器访问http://localhost:6000

# Bugs(todo)

- 多次重复启动一个服务时，依赖于该服务的其他服务收到的服务列表会重复

- gradeService启动后不会收到所依赖的服务更新的通知
