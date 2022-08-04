package regdb

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestInjectBuiltInType(t *testing.T) {
	var db RegDB

	db.RegisterObject(123)
	db.RegisterObject("456")
	db.RegisterObject([]int{1, 2, 3})

	var integer int = 1234567890
	db.InjectObject(&integer)
	if integer != 123 {
		t.Errorf("the value of injected integer error, want: 123, got: %d", integer)
	}

	var str string
	db.InjectObject(&str)
	if str != "456" {
		t.Errorf("the value of injected string error, want: \"456\", got: \"%s\"", str)
	}

	var array []int
	db.InjectObject(&array)
	if !reflect.DeepEqual(array, []int{1, 2, 3}) {
		t.Errorf("the value of injected array error, want: [1, 2, 3], got: %v", array)
	}
}

func TestInjectByName(t *testing.T) {
	var db RegDB
	db.RegisterObjectWithName("123", "123")
	db.RegisterObjectWithName("123", 123)
	db.RegisterObjectWithName("456", 456)

	var integer int
	db.InjectObjectByName("123", &integer)
	if integer != 123 {
		t.Errorf("the value of injected integer error, want: 123, got: %d", integer)
	}
}

func TestInjectFailure(t *testing.T) {
	var db RegDB
	db.RegisterObjectWithName("123", 123)

	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("WOW")
		}
	}()

	var integer int
	db.InjectObjectByName("456", &integer)
}

func TestInjectInterface(t *testing.T) {
	var db RegDB
	db.RegisterObject(bufio.NewReader(bytes.NewBuffer(nil)))
	db.RegisterObjectWithName("writer", bufio.NewWriter(bytes.NewBuffer(nil)))

	var reader io.Reader
	db.InjectObject(&reader)

	var writer io.Writer
	db.InjectObjectByName("writer", &writer)
}

func TestRegisterFields(t *testing.T) {
	var db RegDB

	var out = struct {
		Integer int    `out:"integer"`
		String  string `out:""`
	}{
		Integer: 123,
		String:  "456",
	}
	db.RegisterObjectFields(out, "out")

	var integer int
	db.InjectObject(&integer)
	if integer != 123 {
		t.Errorf("the value of injected integer error, want: 123, got: %d", integer)
	}

	var str string
	db.InjectObject(&str)
	if str != "456" {
		t.Errorf("the value of injected string error, want: \"456\", got: \"%s\"", str)
	}
}

func TestInjectFields(t *testing.T) {
	var db RegDB
	db.RegisterObjectWithName("integer", 123)
	db.RegisterObjectWithName("reader", bytes.NewBufferString("test"))

	needInteger := struct {
		Integer int `inject:"integer"`
	}{}
	db.InjectObjectFields(&needInteger, "inject")
	if needInteger.Integer != 123 {
		t.Errorf("the value of injected integer error, want: 123, got: %d", needInteger.Integer)
	}

	needInterface := struct {
		Reader io.Reader `inject:"reader"`
	}{}
	db.InjectObjectFields(&needInterface, "inject")
	data, err := ioutil.ReadAll(needInterface.Reader)
	if err != nil {
		t.Fatalf("Failed to read data")
	}
	if string(data) != "test" {
		t.Errorf("the value of injected string error, want: \"test\", got: \"%s\"", string(data))
	}

	need := struct {
		Integer int       `inject:"integer"`
		Reader  io.Reader `inject:"reader"`
	}{}
	db.InjectObjectFields(&need, "inject")
	if needInteger.Integer != 123 {
		t.Errorf("the value of injected integer error, want: 123, got: %d", needInteger.Integer)
	}
}

func TestRegisterAndInjectFields(t *testing.T) {
	var db RegDB

	out := struct {
		Reader *bufio.Reader `out:"reader"`
		Writer *bufio.Writer `out:"writer"`
	}{
		Reader: bufio.NewReader(bytes.NewBufferString("read")),
		Writer: bufio.NewWriter(bytes.NewBuffer(nil)),
	}
	db.RegisterObjectFields(out, "out")

	in := struct {
		Reader io.Reader `in:"reader"`
		Writer io.Writer `in:"writer"`
	}{}
	db.InjectObjectFields(&in, "in")

	data, err := ioutil.ReadAll(in.Reader)
	if err != nil {
		t.Fatalf("Failed to read data")
	}
	if string(data) != "read" {
		t.Errorf("want: \"read\", got: \"%s\"", string(data))
	}
}

func TestInjectNestedFields(t *testing.T) {
	var db RegDB
	db.RegisterObjectWithName("integer", 123)

	type Nested struct {
		Integer int `in:"integer"`
	}
	in := struct {
		Nested Nested `in:""`
	}{
		Nested{},
	}

	db.InjectObjectFields(&in, "in")

	if in.Nested.Integer != 123 {
		t.Errorf("the value of injected integer error, want: 123, got: %d", in.Nested.Integer)
	}
}

func TestRegisterAndInjectFunction(t *testing.T) {
	var db RegDB
	var factory = func() io.Reader {
		return bytes.NewBufferString("factory")
	}
	db.RegisterObject(factory)

	var in func() io.Reader
	db.InjectObject(&in)

	reader := in()

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read data")
	}
	if string(data) != "factory" {
		t.Errorf("want: \"factory\", got: \"%s\"", string(data))
	}
}
