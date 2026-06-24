package internal

import (
	"fmt"
	"slices"

	"golang.org/x/tools/go/analysis"
)

var _ analysis.Fact = new(HasFieldOrder)

// HasFieldOrder is a Fact attached to structs listing the field instantiation order.
type HasFieldOrder struct {
	// OrderedList ordered slice with the field names in the struct that are expected to be in order.
	OrderedList []string
}

func NewHasFieldOrder(ol []string) *HasFieldOrder {
	return &HasFieldOrder{OrderedList: ol}
}

func (h HasFieldOrder) AFact() {}

// FieldOrder returns a defensive copy of the expected field order.
func (h HasFieldOrder) FieldOrder() []string {
	return slices.Clone(h.OrderedList)
}

func (h HasFieldOrder) String() string {
	return fmt.Sprintf("HasFieldOrder%s", h.OrderedList)
}
