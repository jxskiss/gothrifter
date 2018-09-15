package reflection

import (
	thrift "github.com/jxskiss/thriftkit/lib/thrift"
	"reflect"
	"unsafe"
)

type sliceDecoder struct {
	sliceType   reflect.Type
	elemType    reflect.Type
	elemDecoder internalDecoder
}

func (decoder *sliceDecoder) decode(ptr unsafe.Pointer, r thrift.Reader) error {
	slice := (*sliceHeader)(ptr)
	slice.Len = 0
	offset := uintptr(0)
	_, length, err := r.ReadListBegin()
	if err != nil {
		return err
	}

	if slice.Cap < length {
		newVal := reflect.MakeSlice(decoder.sliceType, 0, length)
		slice.Data = unsafe.Pointer(newVal.Pointer())
		slice.Cap = length
	}
	for i := 0; i < length; i++ {
		if err := decoder.elemDecoder.decode(unsafe.Pointer(uintptr(slice.Data)+offset), r); err != nil {
			return err
		}
		offset += decoder.elemType.Size()
		slice.Len += 1
	}
	return nil
}

// grow grows the slice s so that it can hold extra more values, allocating
// more capacity if needed. It also returns the old and new slice lengths.
func growOne(slice *sliceHeader, sliceType reflect.Type, elementType reflect.Type) {
	newLen := slice.Len + 1
	if newLen <= slice.Cap {
		slice.Len = newLen
		return
	}
	newCap := slice.Cap
	if newCap == 0 {
		newCap = 1
	} else {
		for newCap < newLen {
			if slice.Len < 1024 {
				newCap += newCap
			} else {
				newCap += newCap / 4
			}
		}
	}
	newVal := reflect.MakeSlice(sliceType, newLen, newCap)
	dst := unsafe.Pointer(newVal.Pointer())
	// copy old array into new array
	originalBytesCount := slice.Len * int(elementType.Size())
	srcSliceHeader := (unsafe.Pointer)(&sliceHeader{slice.Data, originalBytesCount, originalBytesCount})
	dstSliceHeader := (unsafe.Pointer)(&sliceHeader{dst, originalBytesCount, originalBytesCount})
	copy(*(*[]byte)(dstSliceHeader), *(*[]byte)(srcSliceHeader))
	slice.Data = dst
	slice.Len = newLen
	slice.Cap = newCap
}

type sliceEncoder struct {
	sliceType   reflect.Type
	elemType    reflect.Type
	elemEncoder internalEncoder
}

func (encoder *sliceEncoder) encode(ptr unsafe.Pointer, w thrift.Writer) error {
	slice := (*sliceHeader)(ptr)
	if err := w.WriteListBegin(encoder.elemEncoder.thriftType(), slice.Len); err != nil {
		return err
	}
	offset := uintptr(slice.Data)
	var addr unsafe.Pointer
	for i := 0; i < slice.Len; i++ {
		addr = unsafe.Pointer(offset)
		if encoder.elemType.Kind() == reflect.Map {
			addr = unsafe.Pointer((uintptr)(*(*uint64)(addr)))
		}
		if err := encoder.elemEncoder.encode(addr, w); err != nil {
			return err
		}
		offset += encoder.elemType.Size()
	}
	return nil
}

func (encoder *sliceEncoder) thriftType() thrift.Type {
	return thrift.LIST
}
