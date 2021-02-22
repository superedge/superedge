package util

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/superedge/superedge/pkg/application-grid-controller/controller/common"
	"hash"
	"hash/fnv"
)

func GenerateHash(template interface{}) uint64 {
	hasher := fnv.New64()
	DeepHashObject(hasher, template)
	return hasher.Sum64()
}

// DeepHashObject writes specified object to hash using the spew library
// which follows pointers and prints actual values of the nested objects
// ensuring the hash does not change when a pointer changes.
func DeepHashObject(hasher hash.Hash, objectToWrite interface{}) {
	hasher.Reset()
	printer := spew.ConfigState{
		Indent:         " ",
		SortKeys:       true,
		DisableMethods: true,
		SpewKeys:       true,
	}
	printer.Fprintf(hasher, "%#v", objectToWrite)
}

func GetTemplateHash(labels map[string]string) string {
	return labels[common.TemplateHashKey]
}

// GetTemplateName returns template name for workload whose name is gridValues, default is 'DefaultTemplateName'
func GetTemplateName(templates map[string]string, gridValues string, defaultTemplateName string) string {
	if name, exist := templates[gridValues]; !exist {
		return defaultTemplateName
	} else {
		return name
	}
}