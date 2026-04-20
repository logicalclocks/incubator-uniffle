/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package filter

import (
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

// NamespaceFilter dynamically tracks which namespaces match a label selector.
// It watches Namespace objects and maintains a thread-safe set of matching names.
type NamespaceFilter struct {
	mu       sync.RWMutex
	matching sets.String
	selector labels.Selector
	factory  informers.SharedInformerFactory
	informer cache.SharedIndexInformer
	disabled bool
}

// NewNamespaceFilter creates a NamespaceFilter from a comma-separated "key=value" selector string.
// If the selector string is empty, the filter is disabled and Matches() always returns true.
func NewNamespaceFilter(kubeClient kubernetes.Interface, selectorString string) *NamespaceFilter {
	if selectorString == "" {
		return &NamespaceFilter{disabled: true}
	}

	parsed := parseSelector(selectorString)
	factory := informers.NewSharedInformerFactory(kubeClient, 0)
	nsInformer := factory.Core().V1().Namespaces().Informer()

	nf := &NamespaceFilter{
		matching: sets.NewString(),
		selector: parsed,
		factory:  factory,
		informer: nsInformer,
	}

	nsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ns, ok := obj.(*corev1.Namespace)
			if !ok {
				return
			}
			nf.evaluate(ns)
		},
		UpdateFunc: func(_, newObj interface{}) {
			ns, ok := newObj.(*corev1.Namespace)
			if !ok {
				return
			}
			nf.evaluate(ns)
		},
		DeleteFunc: func(obj interface{}) {
			var ns *corev1.Namespace
			switch t := obj.(type) {
			case *corev1.Namespace:
				ns = t
			case cache.DeletedFinalStateUnknown:
				if n, ok := t.Obj.(*corev1.Namespace); ok {
					ns = n
				}
			}
			if ns == nil {
				return
			}
			nf.mu.Lock()
			defer nf.mu.Unlock()
			if nf.matching.Has(ns.Name) {
				klog.V(3).Infof("namespace %q removed from filter", ns.Name)
				nf.matching.Delete(ns.Name)
			}
		},
	})

	return nf
}

// Start starts the namespace informer. Call before WaitForCacheSync.
func (nf *NamespaceFilter) Start(stopCh <-chan struct{}) {
	if nf.disabled {
		return
	}
	nf.factory.Start(stopCh)
}

// HasSynced returns true when the namespace informer cache has synced.
func (nf *NamespaceFilter) HasSynced() bool {
	if nf.disabled {
		return true
	}
	return nf.informer.HasSynced()
}

// Matches returns true if the given namespace name is in the matching set.
// When the filter is disabled (no selector configured), always returns true.
func (nf *NamespaceFilter) Matches(namespace string) bool {
	if nf.disabled {
		return true
	}
	nf.mu.RLock()
	defer nf.mu.RUnlock()
	return nf.matching.Has(namespace)
}

// evaluate checks if a namespace matches the selector and updates the set.
func (nf *NamespaceFilter) evaluate(ns *corev1.Namespace) {
	nf.mu.Lock()
	defer nf.mu.Unlock()
	if nf.selector.Matches(labels.Set(ns.Labels)) {
		if !nf.matching.Has(ns.Name) {
			klog.V(3).Infof("namespace %q matches selector, adding to filter", ns.Name)
			nf.matching.Insert(ns.Name)
		}
	} else {
		if nf.matching.Has(ns.Name) {
			klog.V(3).Infof("namespace %q no longer matches selector, removing from filter", ns.Name)
			nf.matching.Delete(ns.Name)
		}
	}
}

// parseSelector parses a "key1=value1,key2=value2" string into a labels.Selector.
func parseSelector(s string) labels.Selector {
	matchLabels := make(map[string]string)
	for _, pair := range strings.Split(s, ",") {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) != 2 {
			klog.Fatalf("invalid namespace-selector %q: each label must be in key=value format", s)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			klog.Fatalf("invalid namespace-selector %q: each label must be in key=value format", s)
		}
		matchLabels[key] = value
	}
	return labels.SelectorFromSet(matchLabels)
}
