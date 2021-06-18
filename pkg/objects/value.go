package objects

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/fract-lang/fract/pkg/grammar"
)

// TODO: Add []Datas to array string function.

const (
	VALInteger  uint8 = 0
	VALFloat    uint8 = 1
	VALString   uint8 = 2
	VALBoolean  uint8 = 3
	VALFunction uint8 = 4
	VALArray    uint8 = 5
)

// Data instance.
type Data struct {
	Data interface{}
	Type uint8
}

// Get data as string.
func (d Data) String() string {
	switch d.Type {
	case VALFunction:
		return "object.function"
	case VALArray:
		if len(d.Data.([]Data)) == 0 {
			return "[]"
		} else {
			var sb strings.Builder
			sb.WriteByte('[')
			for _, data := range d.Data.([]Data) {
				sb.WriteString(data.Format() + " ")
			}
			return sb.String()[:sb.Len()-1] + "]"
		}
	default:
		if d.Data == nil {
			return "0"
		}
		return d.Data.(string)
	}
}

func (d Data) Format() string {
	data := d.String()
	if d.Type == VALString || d.Type == VALBoolean || d.Type == VALFunction || d.Type == VALArray {
		return data
	}
	if data != grammar.KwNaN {
		if d.Type == VALInteger {
			bigfloat, _ := new(big.Float).SetString(data)
			data = bigfloat.String()
			return data
		}
		b, _ := new(big.Float).SetString(data)
		data = b.String()
		if !strings.Contains(data, ".") {
			data = data + ".0"
		}
	}
	return data
}

// Value intance.
type Value struct {
	Content []Data
	Array   bool
}

func (v *Value) Print() bool {
	if v.Content == nil {
		return false
	}

	if v.Array {
		if len(v.Content) == 0 {
			fmt.Print("[]")
		} else {
			var sb strings.Builder
			sb.WriteByte('[')
			for _, data := range v.Content {
				sb.WriteString(data.Format() + " ")
			}
			fmt.Print(sb.String()[:sb.Len()-1] + "]")
		}
	} else {
		fmt.Print(v.Content[0].Format())
	}
	return true
}
