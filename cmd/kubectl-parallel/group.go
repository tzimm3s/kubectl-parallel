package main

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	k8sYaml "k8s.io/apimachinery/pkg/util/yaml"
)

const defaultLabel string = "parallel/group"

type resourceGroups map[string][][]byte

func groupManifests(manifest io.Reader, label string) (resourceGroups, error) {
	decoder := k8sYaml.NewYAMLOrJSONDecoder(manifest, 4096)

	var obj *unstructured.Unstructured

	groups := make(resourceGroups)

	for {
		err := decoder.Decode(&obj)
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("failed to unmarshal manifest: %s", err)
		}

		if obj == nil {
			break
		}

		val, ok := obj.GetLabels()[label]
		if !ok {
			val = "default"
		}

		var resource map[string]interface{}

		resource, err = runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return nil, err
		}

		rawResource, err := yaml.Marshal(resource)
		if err != nil {
			return nil, err
		}

		groups.insert(val, rawResource)

		obj = nil
	}

	return groups, nil
}

func (r *resourceGroups) insert(key string, value []byte) {
	if _, ok := (*r)[key]; !ok {
		(*r)[key] = [][]byte{value}
	} else {
		(*r)[key] = append((*r)[key], value)
	}
}
