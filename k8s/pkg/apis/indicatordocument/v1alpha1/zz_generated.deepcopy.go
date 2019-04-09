// +build !ignore_autogenerated

/*
Copyright The Kubernetes Authors.

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Alert) DeepCopyInto(out *Alert) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Alert.
func (in *Alert) DeepCopy() *Alert {
	if in == nil {
		return nil
	}
	out := new(Alert)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Indicator) DeepCopyInto(out *Indicator) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Indicator.
func (in *Indicator) DeepCopy() *Indicator {
	if in == nil {
		return nil
	}
	out := new(Indicator)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Indicator) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IndicatorDocument) DeepCopyInto(out *IndicatorDocument) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IndicatorDocument.
func (in *IndicatorDocument) DeepCopy() *IndicatorDocument {
	if in == nil {
		return nil
	}
	out := new(IndicatorDocument)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *IndicatorDocument) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IndicatorDocumentList) DeepCopyInto(out *IndicatorDocumentList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]IndicatorDocument, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IndicatorDocumentList.
func (in *IndicatorDocumentList) DeepCopy() *IndicatorDocumentList {
	if in == nil {
		return nil
	}
	out := new(IndicatorDocumentList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *IndicatorDocumentList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IndicatorDocumentSpec) DeepCopyInto(out *IndicatorDocumentSpec) {
	*out = *in
	out.Product = in.Product
	if in.Indicators != nil {
		in, out := &in.Indicators, &out.Indicators
		*out = make([]IndicatorSpec, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.Layout.DeepCopyInto(&out.Layout)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IndicatorDocumentSpec.
func (in *IndicatorDocumentSpec) DeepCopy() *IndicatorDocumentSpec {
	if in == nil {
		return nil
	}
	out := new(IndicatorDocumentSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IndicatorDocumentStatus) DeepCopyInto(out *IndicatorDocumentStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IndicatorDocumentStatus.
func (in *IndicatorDocumentStatus) DeepCopy() *IndicatorDocumentStatus {
	if in == nil {
		return nil
	}
	out := new(IndicatorDocumentStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IndicatorList) DeepCopyInto(out *IndicatorList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Indicator, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IndicatorList.
func (in *IndicatorList) DeepCopy() *IndicatorList {
	if in == nil {
		return nil
	}
	out := new(IndicatorList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *IndicatorList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IndicatorSpec) DeepCopyInto(out *IndicatorSpec) {
	*out = *in
	out.Alert = in.Alert
	if in.Thresholds != nil {
		in, out := &in.Thresholds, &out.Thresholds
		*out = make([]Threshold, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Documentation != nil {
		in, out := &in.Documentation, &out.Documentation
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	in.Presentation.DeepCopyInto(&out.Presentation)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IndicatorSpec.
func (in *IndicatorSpec) DeepCopy() *IndicatorSpec {
	if in == nil {
		return nil
	}
	out := new(IndicatorSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Layout) DeepCopyInto(out *Layout) {
	*out = *in
	if in.Sections != nil {
		in, out := &in.Sections, &out.Sections
		*out = make([]Section, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Layout.
func (in *Layout) DeepCopy() *Layout {
	if in == nil {
		return nil
	}
	out := new(Layout)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Presentation) DeepCopyInto(out *Presentation) {
	*out = *in
	if in.Labels != nil {
		in, out := &in.Labels, &out.Labels
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Presentation.
func (in *Presentation) DeepCopy() *Presentation {
	if in == nil {
		return nil
	}
	out := new(Presentation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Product) DeepCopyInto(out *Product) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Product.
func (in *Product) DeepCopy() *Product {
	if in == nil {
		return nil
	}
	out := new(Product)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Section) DeepCopyInto(out *Section) {
	*out = *in
	if in.Indicators != nil {
		in, out := &in.Indicators, &out.Indicators
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Section.
func (in *Section) DeepCopy() *Section {
	if in == nil {
		return nil
	}
	out := new(Section)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Threshold) DeepCopyInto(out *Threshold) {
	*out = *in
	if in.Lt != nil {
		in, out := &in.Lt, &out.Lt
		*out = new(float64)
		**out = **in
	}
	if in.Lte != nil {
		in, out := &in.Lte, &out.Lte
		*out = new(float64)
		**out = **in
	}
	if in.Eq != nil {
		in, out := &in.Eq, &out.Eq
		*out = new(float64)
		**out = **in
	}
	if in.Neq != nil {
		in, out := &in.Neq, &out.Neq
		*out = new(float64)
		**out = **in
	}
	if in.Gte != nil {
		in, out := &in.Gte, &out.Gte
		*out = new(float64)
		**out = **in
	}
	if in.Gt != nil {
		in, out := &in.Gt, &out.Gt
		*out = new(float64)
		**out = **in
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Threshold.
func (in *Threshold) DeepCopy() *Threshold {
	if in == nil {
		return nil
	}
	out := new(Threshold)
	in.DeepCopyInto(out)
	return out
}
