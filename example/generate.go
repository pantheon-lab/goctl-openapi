package example

//go:generate goctl api plugin -plugin goctl-openapi="openapi -filename user.openapi.yaml" -api user.api -dir .
//go:generate goctl api plugin -plugin goctl-openapi="openapi -filename web.openapi.yaml" -api web.api -dir .
