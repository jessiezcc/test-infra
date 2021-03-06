/*
Copyright 2019 The Knative Authors

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

package resourcetree

import (
	"container/list"
	"reflect"
	"strings"

	"github.com/knative/test-infra/tools/webhook-apicoverage/coveragecalculator"
)

// ResourceForest represents the top-level forest that contains individual resource trees for top-level resource types and all connected nodes across resource trees.
type ResourceForest struct {
	Version string
	TopLevelTrees map[string]ResourceTree // Key is ResourceTree.ResourceName
	ConnectedNodes map[string]*list.List // Head of the linked list keyed by nodeData.fieldType.pkg + nodeData.fieldType.Name()
}

// getConnectedNodeCoverage calculates the outlined coverage for a Type using ConnectedNodes linkedlist. We traverse through each element in the linkedlist and merge
// coverage data into a single coveragecalculator.TypeCoverage object.
func (r *ResourceForest) getConnectedNodeCoverage(fieldType reflect.Type, fieldRules FieldRules, ignoredFields coveragecalculator.IgnoredFields) (coveragecalculator.TypeCoverage) {
	packageName := strings.Replace(fieldType.PkgPath(), "/", ".", -1)
	coverage := coveragecalculator.TypeCoverage {
		Type: fieldType.Name(),
		Package: packageName,
		Fields : make(map[string]*coveragecalculator.FieldCoverage),
	}

	if value, ok := r.ConnectedNodes[fieldType.PkgPath() + "." + fieldType.Name()]; ok {
		for elem := value.Front(); elem != nil; elem = elem.Next() {
			node := elem.Value.(NodeInterface)
			for field, v := range node.GetData().Children {
				if fieldRules.Apply(field) {
					if _, ok := coverage.Fields[field]; !ok {
						coverage.Fields[field] = &coveragecalculator.FieldCoverage {
							Field: field,
							Ignored: ignoredFields.FieldIgnored(packageName, fieldType.Name(), field),
						}
					}
					// merge values across the list.
					coverage.Fields[field].Merge(v.GetData().Covered, v.getValues())
				}
			}
		}
	}
	return coverage
}