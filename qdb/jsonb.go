package qdb

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"reflect"
)

var (
	ErrWrongJsonBType = errors.New("jsonb scan failed.")
)

type JsonB struct {
	data interface{}
}

// Scan implements the Scanner interface.
func (jb *JsonB) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return jb.scanBytes(src)
	case string:
		return jb.scanBytes([]byte(src))
	case nil:
		jb.data = nil
		return nil
	}

	return ErrWrongJsonBType
}

func (jb *JsonB) scanBytes(src []byte) error {
	err := json.Unmarshal(src, jb.data)
	if err != nil {
		return err
	}

	return nil
}

// Value implements the driver Valuer interface.
func (jb *JsonB) Value() (driver.Value, error) {
	j, err := json.Marshal(jb.data)
	return j, err
}

// JSONB wrapp func
// usage example:
// var data SampleData
// db.QueryRow(sqlstr).Scan(qdb.JSONB(&data))
func JSONB(d interface{}) *JsonB {
	if reflect.TypeOf(d).Kind() != reflect.Ptr {
		panic("error JSONB data, must use pointer")
	}

	return &JsonB{data: d}
}
