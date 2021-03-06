/*
Copyright (c) 2021 The cirocket Authors (Neil Hemming)

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

package rocket

// TaskTypeInfoList is a list of task type info.
type TaskTypeInfoList []TaskTypeInfo

// Len is the number of elements in the collection.
func (list TaskTypeInfoList) Len() int {
	return len(list)
}

// Less reports whether the element with index i
// must sort before the element with index j.
func (list TaskTypeInfoList) Less(i, j int) bool {
	return list[i].Type < list[j].Type
}

// Swap swaps the elements with indexes i and j.
func (list TaskTypeInfoList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}
