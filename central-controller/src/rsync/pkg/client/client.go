// SPDX-License-Identifier: Apache-2.0
// Based on Code: https://github.com/johandry/klient

package client

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/validation"
	"k8s.io/client-go/dynamic"
	"k8s.io/apimachinery/pkg/api/meta"

	resapi "k8s.io/apimachinery/pkg/api/resource"
)

// DefaultValidation default action to validate. If `true` all resources by
// default will be validated.
const DefaultValidation = false

// Client is a kubernetes client, like `kubectl`
type Client struct {
	Clientset        kubernetes.Interface
	DynamicClient    dynamic.Interface
	RestMapper       meta.RESTMapper
	factory          *factory
	validator        validation.Schema
	namespace        string
	enforceNamespace bool
	forceConflicts   bool
	ServerSideApply  bool
}

// Result is an alias for the Kubernetes CLI runtime resource.Result
type Result = resource.Result

// BuilderOptions parameters to create a Resource Builder
type BuilderOptions struct {
	Unstructured  bool
	Validate      bool
	Namespace     string
	LabelSelector string
	FieldSelector string
	All           bool
	AllNamespaces bool
}

// NewBuilderOptions creates a BuilderOptions with the default values for
// the parameters to create a Resource Builder
func NewBuilderOptions() *BuilderOptions {
	return &BuilderOptions{
		Unstructured: true,
		Validate:     true,
	}
}

// NewE creates a kubernetes client, returns an error if fail
func NewE(context, kubeconfig string, ns string) (*Client, error) {
	var namespace string
	var enforceNamespace bool
	var err error
	factory := newFactory(context, kubeconfig)

	// If `true` it will always validate the given objects/resources
	// Unless something different is specified in the NewBuilderOptions
	validator, _ := factory.Validator(DefaultValidation)

	if ns == "" {
		namespace, enforceNamespace, err = factory.ToRawKubeConfigLoader().Namespace()
		if err != nil {
			namespace = v1.NamespaceDefault
			enforceNamespace = true
		}
	} else {
		namespace = ns
		enforceNamespace = false
	}

	clientset, err := factory.KubernetesClientSet()
	if err != nil {
		return nil, err
	}
	if clientset == nil {
		return nil, fmt.Errorf("cannot create a clientset from given context and kubeconfig")
	}

	dynamicClient, err := factory.DynamicClient()
	if err != nil {
		return nil, err
	}
	if dynamicClient == nil {
		return nil, fmt.Errorf("cannot create a dynamic client from given context and kubeconfig")
	}

	restMapper, err := factory.ToRESTMapper()
	if err != nil {
		return nil, err
	}
	if restMapper == nil {
		return nil, fmt.Errorf("cannot create a restMapper from given context and kubeconfig")
	}

	return &Client{
		factory:          factory,
		Clientset:        clientset,
		DynamicClient:    dynamicClient,
		RestMapper:       restMapper,
		validator:        validator,
		namespace:        namespace,
		enforceNamespace: enforceNamespace,
	}, nil
}

// New creates a kubernetes client
func New(context, kubeconfig string, namespace string) *Client {
	client, _ := NewE(context, kubeconfig, namespace)
	return client
}

// Builder creates a resource builder
func (c *Client) builder(opt *BuilderOptions) *resource.Builder {
	validator := c.validator
	namespace := c.namespace

	if opt == nil {
		opt = NewBuilderOptions()
	} else {
		if opt.Validate != DefaultValidation {
			validator, _ = c.factory.Validator(opt.Validate)
		}
		if opt.Namespace != "" {
			namespace = opt.Namespace
		}
	}

	b := c.factory.NewBuilder()
	if opt.Unstructured {
		b = b.Unstructured()
	}

	return b.
		Schema(validator).
		ContinueOnError().
		NamespaceParam(namespace).DefaultNamespace()
}

// ResultForFilenameParam returns the builder results for the given list of files or URLs
func (c *Client) ResultForFilenameParam(filenames []string, opt *BuilderOptions) *Result {
	filenameOptions := &resource.FilenameOptions{
		Recursive: false,
		Filenames: filenames,
	}

	return c.builder(opt).
		FilenameParam(c.enforceNamespace, filenameOptions).
		Flatten().
		Do()
}

// ResultForReader returns the builder results for the given reader
func (c *Client) ResultForReader(r io.Reader, opt *BuilderOptions) *Result {
	return c.builder(opt).
		Stream(r, "").
		Flatten().
		Do()
}

// func (c *Client) ResultForName(opt *BuilderOptions, names ...string) *Result {
// 	return c.builder(opt).
// 		LabelSelectorParam(opt.LabelSelector).
// 		FieldSelectorParam(opt.FieldSelector).
// 		SelectAllParam(opt.All).
// 		AllNamespaces(opt.AllNamespaces).
// 		ResourceTypeOrNameArgs(false, names...).RequireObject(false).
// 		Flatten().
// 		Do()
// }

// ResultForContent returns the builder results for the given content
func (c *Client) ResultForContent(content []byte, opt *BuilderOptions) *Result {
	b := bytes.NewBuffer(content)
	return c.ResultForReader(b, opt)
}

func failedTo(action string, info *resource.Info, err error) error {
	var resKind string
	if info.Mapping != nil {
		resKind = info.Mapping.GroupVersionKind.Kind + " "
	}

	return fmt.Errorf("cannot %s object Kind: %q,	Name: %q, Namespace: %q. %s", action, resKind, info.Name, info.Namespace, err)
}

// IsReachable tests connectivity to the cluster
func (c *Client) IsReachable() error {
	client, err := c.factory.KubernetesClientSet()
	if err != nil {
		return fmt.Errorf("Kubernetes cluster unreachable")
	}
	_, err = client.ServerVersion()
	if err != nil {
		return fmt.Errorf("Kubernetes cluster unreachable")
	}
	return nil
}

// PopulateResourceListV1WithDefValues takes strings of form <resourceName1>=<value1>,<resourceName1>=<value2>
// and returns ResourceList.
func PopulateResourceListV1WithDefValues(spec string) (v1.ResourceList, error) {
	// empty input gets a nil response to preserve generator test expected behaviors
	if spec == "" {
		return nil, nil
	}

	result := v1.ResourceList{}
	resourceStatements := strings.Split(spec, ",")
	for _, resourceStatement := range resourceStatements {
		parts := strings.Split(resourceStatement, "=")
		if len(parts) != 2 {
			return nil, fmt.Errorf("PopulateResourceListV1WithDefValues .. Invalid argument syntax %v, expected <resource>=<value>", resourceStatement)
		}
		resourceName := v1.ResourceName(parts[0])
		resourceQuantity, err := resapi.ParseQuantity(parts[1])
		if err != nil {
			return nil, err
		}
		result[resourceName] = resourceQuantity
	}
	return result, nil
}
