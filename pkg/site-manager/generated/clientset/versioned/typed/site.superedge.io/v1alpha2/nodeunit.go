/*
Copyright 2020 The SuperEdge Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1alpha2

import (
	"context"
	"time"

	v1alpha2 "github.com/superedge/superedge/pkg/site-manager/apis/site.superedge.io/v1alpha2"
	scheme "github.com/superedge/superedge/pkg/site-manager/generated/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// NodeUnitsGetter has a method to return a NodeUnitInterface.
// A group's client should implement this interface.
type NodeUnitsGetter interface {
	NodeUnits() NodeUnitInterface
}

// NodeUnitInterface has methods to work with NodeUnit resources.
type NodeUnitInterface interface {
	Create(ctx context.Context, nodeUnit *v1alpha2.NodeUnit, opts v1.CreateOptions) (*v1alpha2.NodeUnit, error)
	Update(ctx context.Context, nodeUnit *v1alpha2.NodeUnit, opts v1.UpdateOptions) (*v1alpha2.NodeUnit, error)
	UpdateStatus(ctx context.Context, nodeUnit *v1alpha2.NodeUnit, opts v1.UpdateOptions) (*v1alpha2.NodeUnit, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha2.NodeUnit, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha2.NodeUnitList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.NodeUnit, err error)
	NodeUnitExpansion
}

// nodeUnits implements NodeUnitInterface
type nodeUnits struct {
	client rest.Interface
}

// newNodeUnits returns a NodeUnits
func newNodeUnits(c *SiteV1alpha2Client) *nodeUnits {
	return &nodeUnits{
		client: c.RESTClient(),
	}
}

// Get takes name of the nodeUnit, and returns the corresponding nodeUnit object, and an error if there is any.
func (c *nodeUnits) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha2.NodeUnit, err error) {
	result = &v1alpha2.NodeUnit{}
	err = c.client.Get().
		Resource("nodeunits").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of NodeUnits that match those selectors.
func (c *nodeUnits) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha2.NodeUnitList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha2.NodeUnitList{}
	err = c.client.Get().
		Resource("nodeunits").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested nodeUnits.
func (c *nodeUnits) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("nodeunits").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a nodeUnit and creates it.  Returns the server's representation of the nodeUnit, and an error, if there is any.
func (c *nodeUnits) Create(ctx context.Context, nodeUnit *v1alpha2.NodeUnit, opts v1.CreateOptions) (result *v1alpha2.NodeUnit, err error) {
	result = &v1alpha2.NodeUnit{}
	err = c.client.Post().
		Resource("nodeunits").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(nodeUnit).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a nodeUnit and updates it. Returns the server's representation of the nodeUnit, and an error, if there is any.
func (c *nodeUnits) Update(ctx context.Context, nodeUnit *v1alpha2.NodeUnit, opts v1.UpdateOptions) (result *v1alpha2.NodeUnit, err error) {
	result = &v1alpha2.NodeUnit{}
	err = c.client.Put().
		Resource("nodeunits").
		Name(nodeUnit.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(nodeUnit).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *nodeUnits) UpdateStatus(ctx context.Context, nodeUnit *v1alpha2.NodeUnit, opts v1.UpdateOptions) (result *v1alpha2.NodeUnit, err error) {
	result = &v1alpha2.NodeUnit{}
	err = c.client.Put().
		Resource("nodeunits").
		Name(nodeUnit.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(nodeUnit).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the nodeUnit and deletes it. Returns an error if one occurs.
func (c *nodeUnits) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("nodeunits").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *nodeUnits) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("nodeunits").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched nodeUnit.
func (c *nodeUnits) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha2.NodeUnit, err error) {
	result = &v1alpha2.NodeUnit{}
	err = c.client.Patch(pt).
		Resource("nodeunits").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
