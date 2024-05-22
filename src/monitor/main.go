package main

import common "github.com/bromistas/weaver-commons"

func main() {
	temp := common.GetVar("world")
	println(temp)
}
