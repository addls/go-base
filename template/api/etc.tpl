# {{.serviceName}} 配置文件

Name: {{.serviceName}}
Host: {{.host}}
Port: {{.port}}

# 应用配置
App:
  Name: {{.serviceName}}
  Version: 1.0.0
  Env: dev  # dev, test, prod

# 日志配置
Log:
  Mode: console  # console, file, volume
  Level: info    # debug, info, warn, error

# 业务配置
# Database:
#   DataSource: "root:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"

# Redis:
#   Host: localhost:6379
#   Pass: ""
