package multiplex

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

const maxQueuedEvents = 1000

type CacheMux interface {
	Name() string
	// Match check the resource url if need multiplex
	Match(method, URLPath string) bool
	// Watch return a event chan from upstream apiserver
	Watch(bookmark bool, ResourceVersion string) (watch.Interface, error)
	// ListObjects list object from informer store, labels filter in store.ListAll
	// field filter will in downstream
	ListObjects(selector labels.Selector, appendFn cache.AppendFunc) error
}
