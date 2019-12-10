package k8s_test

import (
	"reflect"
	"testing"

	"github.com/joshrwolf/kamp/utils/k8s"
)

var (
	resources = `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx
  labels:
    foo: bar
spec:
  template:
  metadata:
    labels:
      app: nginx
  spec:
    containers:
    - image: nginx:1.17.6
      name: nginx
      ports:
      - containerPort: 80
    - image: nginx
      name: nginx-tagless
      ports:
      - containerPort: 80
---
---
apiVersion: extensions/v1beta2
kind: StatefulSet
metadata:
  name: nginx-deployment
spec:
  template:
  spec:
    containers:
    - image: nginx:latest
      name: nginx-latest
      ports:
      - containerPort: 80
`

	wantImages = []string{"nginx:1.17.6", "nginx", "nginx:latest"}
)

func TestSplitYAML(t *testing.T) {
	got := k8s.UnmarshalYAMLResources(resources)

	gotImages := k8s.GetImages(got)

	if !reflect.DeepEqual(gotImages, wantImages) {
		t.Errorf("SplitYAML() = %v, want %v", got, wantImages)
	}
}
