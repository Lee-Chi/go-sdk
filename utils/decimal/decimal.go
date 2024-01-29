package decimal

import (
	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Decimal struct {
	d   decimal.Decimal
	err error
}

var Zero Decimal = Decimal{
	d:   decimal.Zero,
	err: nil,
}

func NewFromString(v string) Decimal {
	d, err := decimal.NewFromString(v)
	if err != nil {
		return Decimal{
			d:   decimal.Zero,
			err: err,
		}
	}

	return Decimal{
		d:   d,
		err: nil,
	}
}

func NewFromFloat(v float64) Decimal {
	d := decimal.NewFromFloat(v)

	return Decimal{
		d:   d,
		err: nil,
	}
}

func NewFromInt(v int64) Decimal {
	d := decimal.NewFromInt(v)

	return Decimal{
		d:   d,
		err: nil,
	}
}

func NewFromDecimal128(v primitive.Decimal128) Decimal {
	bi, exp, err := v.BigInt()
	if err != nil {
		return Decimal{
			d:   decimal.Zero,
			err: err,
		}
	}

	d := decimal.NewFromBigInt(bi, int32(exp))

	return Decimal{
		d:   d,
		err: nil,
	}
}

func (d Decimal) String() string {
	return d.d.String()
}

func (d Decimal) Float64() float64 {
	f, _ := d.d.Float64()
	return f
}

func (d Decimal) Decimal128() primitive.Decimal128 {
	d128, _ := primitive.ParseDecimal128(d.d.String())
	return d128
}

func (d Decimal) IsZero() bool {
	return d.d.IsZero()
}

func (d Decimal) IsPositive() bool {
	return d.d.IsPositive()
}

func (d Decimal) IsNegative() bool {
	return d.d.IsNegative()
}

func (d Decimal) Equal(other Decimal) bool {
	return d.d.Equal(other.d)
}

func (d Decimal) NotEqual(other Decimal) bool {
	return !d.d.Equal(other.d)
}

func (d Decimal) GreaterThan(other Decimal) bool {
	return d.d.GreaterThan(other.d)
}

func (d Decimal) GreaterThanOrEqual(other Decimal) bool {
	return d.d.GreaterThanOrEqual(other.d)
}

func (d Decimal) LessThan(other Decimal) bool {
	return d.d.LessThan(other.d)
}

func (d Decimal) LessThanOrEqual(other Decimal) bool {
	return d.d.LessThanOrEqual(other.d)
}

func (d Decimal) Add(other Decimal) Decimal {
	return Decimal{
		d:   d.d.Add(other.d),
		err: nil,
	}
}

func (d Decimal) Sub(other Decimal) Decimal {
	return Decimal{
		d:   d.d.Sub(other.d),
		err: nil,
	}
}

func (d Decimal) Mul(other Decimal) Decimal {
	return Decimal{
		d:   d.d.Mul(other.d),
		err: nil,
	}
}

func (d Decimal) Div(other Decimal) Decimal {
	return Decimal{
		d:   d.d.Div(other.d),
		err: nil,
	}
}

func (d Decimal) Mod(other Decimal) Decimal {
	return Decimal{
		d:   d.d.Mod(other.d),
		err: nil,
	}
}

func (d Decimal) Pow(other Decimal) Decimal {
	return Decimal{
		d:   d.d.Pow(other.d),
		err: nil,
	}
}

func (d Decimal) Round(places int32) Decimal {
	return Decimal{
		d:   d.d.Round(places),
		err: nil,
	}
}

func (d Decimal) Floor() Decimal {
	return Decimal{
		d:   d.d.Floor(),
		err: nil,
	}
}

func (d Decimal) Ceil() Decimal {
	return Decimal{
		d:   d.d.Ceil(),
		err: nil,
	}
}

func (d Decimal) Truncate(places int32) Decimal {
	return Decimal{
		d:   d.d.Truncate(places),
		err: nil,
	}
}

func (d Decimal) Abs() Decimal {
	return Decimal{
		d:   d.d.Abs(),
		err: nil,
	}
}

func (d Decimal) Neg() Decimal {
	return Decimal{
		d:   d.d.Neg(),
		err: nil,
	}
}

func (d Decimal) Sign() int {
	return d.d.Sign()
}
