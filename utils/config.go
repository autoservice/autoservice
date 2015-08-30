package utils

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"strconv"
)

var numR = regexp.MustCompile(`^0([xob])`)

func setFieldVal(name string, v reflect.Value, str string) error {
	orgStr := str

	switch v.Type().Kind() {
	case reflect.String:
		v.SetString(str)
	case reflect.Bool:
		switch str {
		case "true", "True", "TRUE":
			v.SetBool(true)
		case "false", "False", "FALSE":
			v.SetBool(false)
		default:
			return fmt.Errorf("(%s) invalid bool: `%s`", name, str)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		base := 10
		if subs := numR.FindStringSubmatch(str); len(subs) > 0 {
			str = str[2:]
			switch subs[1] {
			case "x":
				base = 16
			case "o":
				base = 10
			case "b":
				base = 2
			}
		}
		if n, err := strconv.ParseInt(str, base, v.Type().Bits()); err == nil {
			v.SetInt(n)
		} else {
			return fmt.Errorf("(%s) invalid %v: %v", name, v.Type(), orgStr)
		}
	case reflect.Float32, reflect.Float64:
		if n, err := strconv.ParseFloat(str, v.Type().Bits()); err == nil {
			v.SetFloat(n)
		} else {
			return fmt.Errorf("(%s) invalid %v: %v", name, v.Type(), orgStr)
		}
	case reflect.Ptr:
		ele := reflect.New(v.Type().Elem())
		if err := setFieldVal(name, ele.Elem(), str); err == nil {
			v.Set(ele)
		} else {
			return err
		}
	default:
		return fmt.Errorf("(%s) unsupported type: %v", name, v.Type())
	}
	return nil
}

var envR = regexp.MustCompile(`\${(\w+)}`)
var escapeR = regexp.MustCompile(`\\\$`)

func getFieldVal(name string, fieldT *reflect.StructField) (valStr string, err error) {
	valStr = envR.ReplaceAllStringFunc(fieldT.Tag.Get("env"), func(env string) string {
		return os.Getenv(env[1:])
	})
	valStr = escapeR.ReplaceAllString(valStr, `$`)
	if valStr == "" {
		valStr = fieldT.Tag.Get("default")
	}
	return valStr, nil
}

func InitConfig(typ reflect.Type, val reflect.Value, name string) error {
	if typ.Kind() == reflect.Ptr {
		tmpT := typ
		for tmpT.Kind() == reflect.Ptr {
			tmpT = tmpT.Elem()
			if val.IsNil() {
				val.Set(reflect.New(tmpT))
			}
			val = val.Elem()
		}
	}
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("(%s) invalid config type: %v", name, val.Type())
	}
	if name == "" {
		name = val.Type().Name()
	}

	vtype := val.Type()
	for i := 0; i < vtype.NumField(); i++ {
		field := vtype.Field(i)
		fieldName := name + "." + field.Name
		if field.Anonymous {
			if err := InitConfig(field.Type, val.Field(i), fieldName); err != nil {
				return err
			}
			continue
		}

		if field.Type.Kind() == reflect.Ptr || field.Type.Kind() == reflect.Struct {
			tmpT := field.Type
			for tmpT.Kind() == reflect.Ptr {
				tmpT = tmpT.Elem()
			}
			if tmpT.Kind() == reflect.Struct {
				if err := InitConfig(field.Type, val.Field(i), fieldName); err != nil {
					return err
				}
				continue
			}
		}

		if valStr, err := getFieldVal(fieldName, &field); err != nil {
			return err
		} else if valStr != "" {
			if err := setFieldVal(fieldName, val.Field(i), valStr); err != nil {
				return err
			}
		}
	}

	return nil
}

func LoadConfig(v interface{}, data []byte) error {
	if err := InitConfig(reflect.TypeOf(v), reflect.ValueOf(v), ""); err != nil {
		return err
	}

	return yaml.Unmarshal(data, v)
}

func LoadConfigFromFile(v interface{}, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	return LoadConfig(v, data)
}
