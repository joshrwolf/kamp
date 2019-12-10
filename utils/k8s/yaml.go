package k8s

import (
	"fmt"
	"regexp"

	"github.com/ghodss/yaml"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	log "github.com/sirupsen/logrus"
)

// Image simply strictly types Images with strings
type Image = string

var yamlSeparator = regexp.MustCompile(`\n---`)

// UnmarshalYAMLResources takes in a string YAML representation and returns a list of metav1.unstructured type objects
func UnmarshalYAMLResources(data string) (objs []*unstructured.Unstructured) {
	// Split the yaml data into parts based off the yaml sep
	parts := yamlSeparator.Split(data, -1)

	// Loop over each part and unmarshal
	for _, part := range parts {
		// Check for empty yaml
		if len(part) == 0 {
			log.Debugf("Empty manifests found, skipping...")
			continue
		}

		var obj unstructured.Unstructured
		err := yaml.Unmarshal([]byte(part), &obj)
		if err != nil {
			log.Fatalf("Failed to unmarshal map into unstructured: %v", err)
		}

		objs = append(objs, &obj)
	}
	return
}

// GetImages returns a slice of images (type Image = string) found within a unstructured objs list
func GetImages(objs []*unstructured.Unstructured) (images []Image) {
	for _, obj := range objs {
		images = append(images, getImage(obj.Object)...)
	}
	return
}

func getImage(obj map[string]interface{}) (images []Image) {
	// Loop through every yaml obj
	for k, v := range obj {
		if array, ok := v.([]interface{}); ok {
			// *containers are the only things that have images so they're the only things we care about... duh
			if k == "containers" || k == "initContainers" {
				for _, obj := range array {
					if mapObj, mapOk := obj.(map[string]interface{}); mapOk {
						if image, isImage := mapObj["image"]; isImage {
							images = append(images, fmt.Sprintf("%s", image))
						}
					}
				}
			}
		} else if objMap, ok := v.(map[string]interface{}); ok {
			// Keep going till we run out of maps or find a container
			images = append(images, getImage(objMap)...)
		}
	}
	return
}
