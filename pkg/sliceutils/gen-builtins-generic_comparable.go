// Code generated by genny. DO NOT EDIT.
// This file was automatically generated by genny.
// Any changes will be lost if this file is regenerated.
// see https://github.com/mauricelam/genny

package sliceutils

import (
	"sort"
)

type sortableFloat32Slice []float32

func (s sortableFloat32Slice) Len() int {
	return len(s)
}

func (s sortableFloat32Slice) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s sortableFloat32Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Float32Sort sorts the given slice.
func Float32Sort(slice []float32) {
	sort.Sort(sortableFloat32Slice(slice))
}

type sortableFloat64Slice []float64

func (s sortableFloat64Slice) Len() int {
	return len(s)
}

func (s sortableFloat64Slice) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s sortableFloat64Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Float64Sort sorts the given slice.
func Float64Sort(slice []float64) {
	sort.Sort(sortableFloat64Slice(slice))
}

type sortableIntSlice []int

func (s sortableIntSlice) Len() int {
	return len(s)
}

func (s sortableIntSlice) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s sortableIntSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// IntSort sorts the given slice.
func IntSort(slice []int) {
	sort.Sort(sortableIntSlice(slice))
}

type sortableInt16Slice []int16

func (s sortableInt16Slice) Len() int {
	return len(s)
}

func (s sortableInt16Slice) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s sortableInt16Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Int16Sort sorts the given slice.
func Int16Sort(slice []int16) {
	sort.Sort(sortableInt16Slice(slice))
}

type sortableInt32Slice []int32

func (s sortableInt32Slice) Len() int {
	return len(s)
}

func (s sortableInt32Slice) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s sortableInt32Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Int32Sort sorts the given slice.
func Int32Sort(slice []int32) {
	sort.Sort(sortableInt32Slice(slice))
}

type sortableInt64Slice []int64

func (s sortableInt64Slice) Len() int {
	return len(s)
}

func (s sortableInt64Slice) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s sortableInt64Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Int64Sort sorts the given slice.
func Int64Sort(slice []int64) {
	sort.Sort(sortableInt64Slice(slice))
}

type sortableInt8Slice []int8

func (s sortableInt8Slice) Len() int {
	return len(s)
}

func (s sortableInt8Slice) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s sortableInt8Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Int8Sort sorts the given slice.
func Int8Sort(slice []int8) {
	sort.Sort(sortableInt8Slice(slice))
}

type sortableUintSlice []uint

func (s sortableUintSlice) Len() int {
	return len(s)
}

func (s sortableUintSlice) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s sortableUintSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// UintSort sorts the given slice.
func UintSort(slice []uint) {
	sort.Sort(sortableUintSlice(slice))
}

type sortableUint16Slice []uint16

func (s sortableUint16Slice) Len() int {
	return len(s)
}

func (s sortableUint16Slice) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s sortableUint16Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Uint16Sort sorts the given slice.
func Uint16Sort(slice []uint16) {
	sort.Sort(sortableUint16Slice(slice))
}

type sortableUint32Slice []uint32

func (s sortableUint32Slice) Len() int {
	return len(s)
}

func (s sortableUint32Slice) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s sortableUint32Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Uint32Sort sorts the given slice.
func Uint32Sort(slice []uint32) {
	sort.Sort(sortableUint32Slice(slice))
}

type sortableUint64Slice []uint64

func (s sortableUint64Slice) Len() int {
	return len(s)
}

func (s sortableUint64Slice) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s sortableUint64Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Uint64Sort sorts the given slice.
func Uint64Sort(slice []uint64) {
	sort.Sort(sortableUint64Slice(slice))
}

type sortableUint8Slice []uint8

func (s sortableUint8Slice) Len() int {
	return len(s)
}

func (s sortableUint8Slice) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s sortableUint8Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Uint8Sort sorts the given slice.
func Uint8Sort(slice []uint8) {
	sort.Sort(sortableUint8Slice(slice))
}

type sortableStringSlice []string

func (s sortableStringSlice) Len() int {
	return len(s)
}

func (s sortableStringSlice) Less(i, j int) bool {
	return s[i] < s[j]
}

func (s sortableStringSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// StringSort sorts the given slice.
func StringSort(slice []string) {
	sort.Sort(sortableStringSlice(slice))
}
