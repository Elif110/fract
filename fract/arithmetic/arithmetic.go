package arithmetic

import (
	"fmt"
	"regexp"
	"strconv"

	"../../fract"
	"../../grammar"
)

// IsTypesCompatible Check types are compatible?
// type0 Primary type.
// type1 Secondary type.
func IsTypesCompatible(type0 int, type1 int) bool {
	if IsIntegerType(type0) {
		return IsIntegerType(type1)
	}
	return IsFloatType(type1)
}

// IsIntegerType Type is integer?
// _type Type to check.
func IsIntegerType(_type int) bool {
	return _type == fract.TypeShort ||
		_type == fract.TypeInt ||
		_type == fract.TypeLong ||
		_type == fract.TypeUShort ||
		_type == fract.TypeUInt ||
		_type == fract.TypeULong
}

// IsFloatType Type is float?
// _type Type to check.
func IsFloatType(_type int) bool {
	return _type == fract.TypeFloat ||
		_type == fract.TypeDouble
}

// IsNegative Is negative number?
// value Value to check.
func IsNegative(value string) bool {
	return value[0] == grammar.TokenMinus[0]
}

// IsNumeric Char is numeric?
// char Char to check.
func IsNumeric(char byte) bool {
	return char == '0' ||
		char == '1' ||
		char == '2' ||
		char == '3' ||
		char == '4' ||
		char == '5' ||
		char == '6' ||
		char == '7' ||
		char == '8' ||
		char == '9'
}

// IsInteger Value is an integer?
// value Value to check.
func IsInteger(value string) bool {
	state, _ := regexp.MatchString("^(-|)\\s*[0-9]+$", value)
	return state
}

// IsFloat Value is an float?
// value Value to check.
func IsFloat(value string) bool {
	state, _ := regexp.MatchString("^(-|)\\s*[0-9]+(\\.[0-9]+)?$", value)
	return state
}

// ToFloat String to float.
// value Value to parse.
func ToFloat(value string) (float32, error) {
	result, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return 0, err
	}
	return float32(result), err
}

// ToDouble String to double.
// value Value to parse.
func ToDouble(value string) (float64, error) {
	return strconv.ParseFloat(value, 64)
}

// ToSByte String to 8bit integer.
// value Value to parse.
func ToSByte(value string) (int8, error) {
	result, err := strconv.ParseInt(value, 10, 8)
	if err != nil {
		return 0, err
	}
	return int8(result), err
}

// ToShort String to 16bit integer.
// value Value to parse.
func ToShort(value string) (int16, error) {
	result, err := strconv.ParseInt(value, 10, 16)
	if err != nil {
		return 0, err
	}
	return int16(result), err
}

// ToInt String to 32bit integer.
// value Value to parse.
func ToInt(value string) (int32, error) {
	result, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(result), err
}

// ToLong String to 64bit integer.
// value Value to parse.
func ToLong(value string) (int64, error) {
	result, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return int64(result), err
}

// ToByte String to 8bit unsigned integer.
// value Value to parse.
func ToByte(value string) (uint8, error) {
	result, err := strconv.ParseUint(value, 10, 8)
	if err != nil {
		return 0, err
	}
	return uint8(result), err
}

// ToUShort String to 16bit unsigned integer.
// value Value to parse.
func ToUShort(value string) (uint16, error) {
	result, err := strconv.ParseUint(value, 10, 16)
	if err != nil {
		return 0, err
	}
	return uint16(result), err
}

// ToUInt String to 32bit integer.
// value Value to parse.
func ToUInt(value string) (uint32, error) {
	result, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(result), err
}

// ToULong String to 64bit unsigned integer.
// value Value to parse.
func ToULong(value string) (uint64, error) {
	result, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint64(result), err
}

// FloatToString Float to string.
// value Value to parse.
func FloatToString(value float64) string {
	return fmt.Sprintf("%f", value)
}

// IntToString Integer to string.
// value Value to parse.
func IntToString(value int) string {
	return fmt.Sprintf("%d", value)
}

// ByteToString Byte to string.
// value Value to parse.
func ByteToString(value byte) string {
	return fmt.Sprintf("%d", value)
}
