module github.com/go-spring/starter-redigo

go 1.14

require (
	github.com/go-spring/spring-base v1.1.0-rc2
	github.com/go-spring/spring-core v1.1.0-rc2
	github.com/go-spring/spring-redigo v1.1.0-rc2
)

replace (
	github.com/go-spring/spring-core => ../../spring/spring-core
	github.com/go-spring/spring-redigo => ../../spring/spring-redigo
)
