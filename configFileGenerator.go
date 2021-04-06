package main

import (
	"gopkg.in/yaml.v2"
)

type Image struct {
	Repository string
	Tag        string
}

type Ingress struct {
	Path string
}

type HelmValuesFile struct {
	ReplicaCount     int
	FullnameOverride string

	Image   Image
	Ingress Ingress
}

func CreateHelmValuesFile(service Service) []byte {
	helmValuesFile := HelmValuesFile{
		ReplicaCount:     1,
		FullnameOverride: service.Name,

		Image: Image{
			Repository: service.ContainerRepository,
			Tag:        service.ImageTag,
		},

		Ingress: Ingress{
			Path: "/" + service.GatewayEndpoint,
		},
	}

	bytes, err := yaml.Marshal(helmValuesFile)
	PanicIfError(err)

	return bytes
}

type ArgoAppValuesFile struct {
	Name            string
	GatewayEndpoint string
}

func CreateArgoAppFile(service Service) []byte {
	argoAppValuesFile := ArgoAppValuesFile{
		Name:            service.Name,
		GatewayEndpoint: service.GatewayEndpoint,
	}

	bytes, err := yaml.Marshal(argoAppValuesFile)
	PanicIfError(err)

	return bytes
}
