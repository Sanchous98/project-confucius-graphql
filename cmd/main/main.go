package main

import (
	confucius "github.com/Sanchous98/project-confucius-base"
	"github.com/Sanchous98/project-confucius-graphql/src"
)

func main() {
	confucius.App().Bind(&src.GraphQL{}).Launch(true)
}
