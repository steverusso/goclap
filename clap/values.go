package clap

import (
	"flag"
	"fmt"
	"strconv"
)

var (
	_ = (flag.Value)((*Bool)(nil))
	_ = (flag.Value)((*String)(nil))
	_ = (flag.Value)((*Float32)(nil))
	_ = (flag.Value)((*Float64)(nil))

	_ = (flag.Value)((*Int)(nil))
	_ = (flag.Value)((*Int8)(nil))
	_ = (flag.Value)((*Int16)(nil))
	_ = (flag.Value)((*Int32)(nil))
	_ = (flag.Value)((*Int64)(nil))

	_ = (flag.Value)((*Uint)(nil))
	_ = (flag.Value)((*Uint8)(nil))
	_ = (flag.Value)((*Uint16)(nil))
	_ = (flag.Value)((*Uint32)(nil))
	_ = (flag.Value)((*Uint64)(nil))
)

type Bool bool

func NewBool(p *bool) *Bool { return (*Bool)(p) }

func (v *Bool) String() string { return strconv.FormatBool(bool(*v)) }

func (v *Bool) Set(s string) error {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return fmt.Errorf(`invalid boolean value "%s"`, s)
	}
	*v = Bool(b)
	return err
}

func (*Bool) IsBoolFlag() bool { return true }

type String string

func NewString(p *string) *String { return (*String)(p) }

func (v *String) String() string { return string(*v) }

func (v *String) Set(s string) error {
	*v = String(s)
	return nil
}

type Float32 float32

func NewFloat32(p *float32) *Float32 { return (*Float32)(p) }

func (v *Float32) String() string { return strconv.FormatFloat(float64(*v), 'g', -1, 32) }

func (v *Float32) Set(s string) error {
	f64, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return numError(err)
	}
	*v = Float32(f64)
	return err
}

type Float64 float64

func NewFloat64(p *float64) *Float64 { return (*Float64)(p) }

func (v *Float64) String() string { return strconv.FormatFloat(float64(*v), 'g', -1, 64) }

func (v *Float64) Set(s string) error {
	f64, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return numError(err)
	}
	*v = Float64(f64)
	return err
}

type Int int

func NewInt(p *int) *Int { return (*Int)(p) }

func (v *Int) String() string { return strconv.Itoa(int(*v)) }

func (v *Int) Set(s string) error {
	i64, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err != nil {
		return numError(err)
	}
	*v = Int(i64)
	return err
}

type Int8 int8

func NewInt8(p *int8) *Int8 { return (*Int8)(p) }

func (v *Int8) String() string { return strconv.FormatInt(int64(*v), 10) }

func (v *Int8) Set(s string) error {
	i64, err := strconv.ParseInt(s, 0, 8)
	if err != nil {
		return numError(err)
	}
	*v = Int8(i64)
	return nil
}

type Int16 int16

func NewInt16(p *int16) *Int16 { return (*Int16)(p) }

func (v *Int16) String() string { return strconv.FormatInt(int64(*v), 10) }

func (v *Int16) Set(s string) error {
	i64, err := strconv.ParseInt(s, 0, 16)
	if err != nil {
		return numError(err)
	}
	*v = Int16(i64)
	return nil
}

type Int32 int32

func NewInt32(p *int32) *Int32 { return (*Int32)(p) }

func (v *Int32) String() string { return strconv.FormatInt(int64(*v), 10) }

func (v *Int32) Set(s string) error {
	i64, err := strconv.ParseInt(s, 0, 32)
	if err != nil {
		return numError(err)
	}
	*v = Int32(i64)
	return nil
}

type Int64 int64

func NewInt64(p *int64) *Int64 { return (*Int64)(p) }

func (v *Int64) String() string { return strconv.FormatInt(int64(*v), 10) }

func (v *Int64) Set(s string) error {
	i64, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		return numError(err)
	}
	*v = Int64(i64)
	return nil
}

type Uint uint

func NewUint(p *uint) *Uint { return (*Uint)(p) }

func (v *Uint) String() string { return strconv.FormatUint(uint64(*v), 10) }

func (v *Uint) Set(s string) error {
	u64, err := strconv.ParseUint(s, 0, strconv.IntSize)
	if err != nil {
		return numError(err)
	}
	*v = Uint(u64)
	return err
}

type Uint8 uint8

func NewUint8(p *uint8) *Uint8 { return (*Uint8)(p) }

func (v *Uint8) String() string { return strconv.FormatUint(uint64(*v), 10) }

func (v *Uint8) Set(s string) error {
	u64, err := strconv.ParseUint(s, 0, 8)
	if err != nil {
		return numError(err)
	}
	*v = Uint8(u64)
	return nil
}

type Uint16 uint16

func NewUint16(p *uint16) *Uint16 { return (*Uint16)(p) }

func (v *Uint16) String() string { return strconv.FormatUint(uint64(*v), 10) }

func (v *Uint16) Set(s string) error {
	u64, err := strconv.ParseUint(s, 0, 16)
	if err != nil {
		return numError(err)
	}
	*v = Uint16(u64)
	return nil
}

type Uint32 uint32

func NewUint32(p *uint32) *Uint32 { return (*Uint32)(p) }

func (v *Uint32) String() string { return strconv.FormatUint(uint64(*v), 10) }

func (v *Uint32) Set(s string) error {
	u64, err := strconv.ParseUint(s, 0, 32)
	if err != nil {
		return numError(err)
	}
	*v = Uint32(u64)
	return nil
}

type Uint64 uint64

func NewUint64(p *uint64) *Uint64 { return (*Uint64)(p) }

func (v *Uint64) String() string { return strconv.FormatUint(uint64(*v), 10) }

func (v *Uint64) Set(s string) error {
	u64, err := strconv.ParseUint(s, 0, 64)
	if err != nil {
		return numError(err)
	}
	*v = Uint64(u64)
	return nil
}

func numError(err error) error {
	if ne, ok := err.(*strconv.NumError); ok {
		return ne.Err
	}
	return err
}
