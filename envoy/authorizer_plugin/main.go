//go:build authorizer_plugin

package main

import (
	_ "github.com/buildwithgrove/authorizer-plugin/db"
	_ "github.com/buildwithgrove/authorizer-plugin/db/postgres"
	_ "github.com/buildwithgrove/authorizer-plugin/filter"
	_ "github.com/buildwithgrove/authorizer-plugin/user"
)

func main() {}
