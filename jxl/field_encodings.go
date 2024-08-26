package jxl

import "gojxl/jxl/base"

type Fields interface {
	VisitFields(visitor Visitor) base.Status
}
