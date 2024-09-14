package main

import (
	_ "github.com/buildwithgrove/path-authorizer/db"
	_ "github.com/buildwithgrove/path-authorizer/db/postgres"
	_ "github.com/buildwithgrove/path-authorizer/filter"
	_ "github.com/buildwithgrove/path-authorizer/filter/handler"
	_ "github.com/buildwithgrove/path-authorizer/user"
)

func main() {}
