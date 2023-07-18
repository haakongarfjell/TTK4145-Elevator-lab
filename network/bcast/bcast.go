package bcast

import (
	"elevator-project/network/conn"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
)

const buf_size = 1024

// Encodes received values from `chans` into type-tagged JSON, then broadcasts
// it on `port`
func Transmitter(port int, chans ...interface{}) {
	checkArgs(chans...)
	type_names := make([]string, len(chans))
	select_cases := make([]reflect.SelectCase, len(type_names))
	for i, ch := range chans {
		select_cases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ch),
		}
		type_names[i] = reflect.TypeOf(ch).Elem().String()
	}

	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp4", fmt.Sprintf("255.255.255.255:%d", port))
	for {
		chosen, value, _ := reflect.Select(select_cases)
		jsonstr, _ := json.Marshal(value.Interface())
		ttj, _ := json.Marshal(typeTaggedJSON{
			TypeId: type_names[chosen],
			JSON:   jsonstr,
		})
		if len(ttj) > buf_size {
			panic(fmt.Sprintf(
				"Tried to send a message longer than the buffer size (length: %d, buffer size: %d)\n\t'%s'\n"+
					"Either send smaller packets, or go to network/bcast/bcast.go and increase the buffer size",
				len(ttj), buf_size, string(ttj)))
		}
		conn.WriteTo(ttj, addr)

	}
}

// Matches type-tagged JSON received on `port` to element types of `chans`, then
// sends the decoded value on the corresponding channel
func Receiver(port int, chans ...interface{}) {
	checkArgs(chans...)
	chans_map := make(map[string]interface{})
	for _, ch := range chans {
		chans_map[reflect.TypeOf(ch).Elem().String()] = ch
	}

	var buf [buf_size]byte
	conn := conn.DialBroadcastUDP(port)
	for {
		n, _, e := conn.ReadFrom(buf[0:])
		if e != nil {
			fmt.Printf("bcast.Receiver(%d, ...):ReadFrom() failed: \"%+v\"\n", port, e)
		}

		var ttj typeTaggedJSON
		json.Unmarshal(buf[0:n], &ttj)
		ch, ok := chans_map[ttj.TypeId]
		if !ok {
			continue
		}
		v := reflect.New(reflect.TypeOf(ch).Elem())
		json.Unmarshal(ttj.JSON, v.Interface())
		reflect.Select([]reflect.SelectCase{{
			Dir:  reflect.SelectSend,
			Chan: reflect.ValueOf(ch),
			Send: reflect.Indirect(v),
		}})
	}
}

type typeTaggedJSON struct {
	TypeId string
	JSON   []byte
}

// Checks that args to Tx'er/Rx'er are valid:
//
//	All args must be channels
//	Element types of channels must be encodable with JSON
//	No element types are repeated
//
// Implementation note:
//   - Why there is no `isMarshalable()` function in encoding/json is a mystery,
//     so the tests on element type are hand-copied from `encoding/json/encode.go`
func checkArgs(chans ...interface{}) {
	n := 0
	for range chans {
		n++
	}
	elem_types := make([]reflect.Type, n)

	for i, ch := range chans {
		// Must be a channel
		if reflect.ValueOf(ch).Kind() != reflect.Chan {
			panic(fmt.Sprintf(
				"Argument must be a channel, got '%s' instead (arg# %d)",
				reflect.TypeOf(ch).String(), i+1))
		}

		elem_type := reflect.TypeOf(ch).Elem()

		// Element type must not be repeated
		for j, e := range elem_types {
			if e == elem_type {
				panic(fmt.Sprintf(
					"All channels must have mutually different element types, arg# %d and arg# %d both have element type '%s'",
					j+1, i+1, e.String()))
			}
		}
		elem_types[i] = elem_type

		// Element type must be encodable with JSON
		checkTypeRecursive(elem_type, []int{i + 1})

	}
}

func checkTypeRecursive(val reflect.Type, offsets []int) {
	switch val.Kind() {
	case reflect.Complex64, reflect.Complex128, reflect.Chan, reflect.Func, reflect.UnsafePointer:
		panic(fmt.Sprintf(
			"Channel element type must be supported by JSON, got '%s' instead (nested arg# %v)",
			val.String(), offsets))
	case reflect.Map:
		if val.Key().Kind() != reflect.String {
			panic(fmt.Sprintf(
				"Channel element type must be supported by JSON, got '%s' instead (map keys must be 'string') (nested arg# %v)",
				val.String(), offsets))
		}
		checkTypeRecursive(val.Elem(), offsets)
	case reflect.Array, reflect.Ptr, reflect.Slice:
		checkTypeRecursive(val.Elem(), offsets)
	case reflect.Struct:
		for idx := 0; idx < val.NumField(); idx++ {
			checkTypeRecursive(val.Field(idx).Type, append(offsets, idx+1))
		}
	}
}