package qdb

import (
	"bytes"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/lib/pq"
)

var (
	ErrWrongPointType = errors.New("piont scan failed")
)

type Point struct {
	X float64
	Y float64
}

// Scan implements the Scanner interface.
func (pt *Point) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return pt.scanBytes(src)
	case string:
		return pt.scanBytes([]byte(src))
	case nil:
		pt = nil
		return nil
	}
	return ErrWrongPointType

}

func (pt *Point) scanBytes(src []byte) error {
	// 去掉括号并按逗号分割字符串
	parts := strings.Split(strings.Trim(string(src), "\"()"), ",")

	if len(parts) != 2 {
		return ErrWrongPointType
	}

	var err error
	pt.X, err = strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return ErrWrongPointType
	}

	pt.Y, err = strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return ErrWrongPointType
	}

	return nil
}

// Value implements the driver Valuer interface.
func (pt *Point) Value() (driver.Value, error) {
	buf := new(bytes.Buffer)
	fmt.Fprintf(buf, "(%f, %f)", pt.X, pt.Y)
	return buf.Bytes(), nil
}

type PointArray []Point

// Scan implements the Scanner interface.
func (a *PointArray) Scan(src interface{}) error {
	var err error

	var elems pq.StringArray

	if err = elems.Scan(src); err != nil {
		return err
	}

	if *a != nil && len(elems) == 0 {
		*a = (*a)[:0]
	} else {
		b := make(PointArray, len(elems))
		for i, v := range elems {
			if err = b[i].Scan(v); err != nil {
				return err
			}
		}
		*a = b
	}

	return nil
}

// Value implements the driver Valuer interface.
func (a PointArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	if n := len(a); n > 0 {
		buf := new(bytes.Buffer)
		buf.WriteString("{")
		fmt.Fprintf(buf, "\"(%f,%f)\"", a[0].X, a[0].Y)
		for i := 1; i < n; i++ {
			buf.WriteByte(',')
			fmt.Fprintf(buf, "\"(%f,%f)\"", a[i].X, a[i].Y)
		}

		buf.WriteString("}")

		return buf.Bytes(), nil
	}
	return "{}", nil
}
