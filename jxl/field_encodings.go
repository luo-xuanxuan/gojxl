package jxl

type Fields interface {
	VisitFields(visitor Visitor) Status
}
