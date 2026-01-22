package server

import (
	{{.importPackages}}
)

type {{.service}}Server struct {
	svcCtx *svc.ServiceContext
	{{.unimplementedServer}}
}

func New{{.service}}Server(svcCtx *svc.ServiceContext) *{{.service}}Server {
	return &{{.service}}Server{
		svcCtx: svcCtx,
	}
}
