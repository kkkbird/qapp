package qdb

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"reflect"
)

var (
	ErrWrongJsonBType = errors.New("Type assertion .([]byte) failed.")
)

type JsonB struct {
	data interface{}
}

func (jb *JsonB) Scan(src interface{}) error {
	source, ok := src.([]byte)
	if !ok {
		return ErrWrongJsonBType
	}

	err := json.Unmarshal(source, jb.data)
	if err != nil {
		return err
	}

	return nil
}

func (jb *JsonB) Value() (driver.Value, error) {
	j, err := json.Marshal(jb.data)
	return j, err
}

func JSONB(d interface{}) *JsonB {
	if reflect.TypeOf(d).Kind() != reflect.Ptr {
		panic("error JSONB data, must use pointer")
	}

	return &JsonB{data: d}
}
