package iniconfig

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
)

// 从文件反序列化
func UnMarshalFile(filename string, i interface{}) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	return UnMarshal(data, i)
}

// 序列化到文件
func MarshalFile(filename string, i interface{}) error {

	data, err := Marshal(i)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

// 反序列化 ini配置文件
func UnMarshal(data []byte, v interface{}) error {
	var lastSectionName string
	s := strings.Split(string(data), "\n")

	t := reflect.TypeOf(v)

	if t.Kind() != reflect.Ptr {
		err := errors.New("please pass address")
		return err
	}

	if t.Elem().Kind() != reflect.Struct {
		err := errors.New("please pass struct")
		return err
	}

	for i, line := range s {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// 忽略注释
		if line[0] == ';' || line[0] == '#' {
			continue
		}

		if line[0] == '[' {
			SectionName, err := parseSection(line, t)

			if err != nil {
				err = fmt.Errorf("error:%v lineno:%d", err, i+1)
				return err
			}
			lastSectionName = SectionName
			continue
		}

		err := parseItem(line, lastSectionName, v)
		if err != nil {
			err = fmt.Errorf("%v lineno:%d", err, i+1)
			return err
		}
	}

	return nil
}

// 解析节点[server]
func parseSection(line string, t reflect.Type) (string, error) {
	var lastSectionName string
	if len(line) <= 2 || line[len(line)-1] != ']' {
		err := fmt.Errorf("syntax error,invalid section:%s", line)
		return "", err
	}

	sectionName := strings.TrimSpace(line[1 : len(line)-1])
	if len(sectionName) == 0 {
		err := fmt.Errorf("syntax error,invalid section:%s", line)
		return "", err
	}

	for i := 0; i < t.Elem().NumField(); i++ {
		filed := t.Elem().Field(i)
		tagVal := filed.Tag.Get("ini")
		if tagVal == sectionName {
			lastSectionName = filed.Name
			break
		}
	}

	return lastSectionName, nil
}

// 解析字段 xxx = xxxx
func parseItem(line, lastSectionName string, i interface{}) error {
	var filedKeyName string
	// 如果节点还没有获取到直接报错
	if len(lastSectionName) == 0 {
		err := fmt.Errorf("syntax err,please check config file")
		return err
	}

	// 获取=号索引
	index := strings.Index(line, "=")

	// 如果没有找到=号直接报语法错误
	if index == -1 {
		err := fmt.Errorf("syntax error line:%s", line)
		return err
	}

	// 获取key和val ip = 10.0.0.1
	key := strings.TrimSpace(line[:index])
	val := strings.TrimSpace(line[index+1:])

	// 如果没写key，返回错误
	if len(key) == 0 {
		err := fmt.Errorf("syntax error, line: %s", line)
		return err
	}

	// 节点的值信息
	sectionVal := reflect.ValueOf(i).Elem().FieldByName(lastSectionName)
	sectionType := sectionVal.Type()
	// 如果节点不是结构体，直接报错
	if sectionType.Kind() != reflect.Struct {
		err := fmt.Errorf("filed %s must be a struct", sectionVal.Elem().FieldByName(lastSectionName))
		return err
	}

	if sectionVal == reflect.ValueOf(nil) {
		return fmt.Errorf("no filed")
	}

	// 循环获取节点下的字段的字段
	for i := 0; i < sectionType.NumField(); i++ {
		// 获取字段的ini名
		filedName := sectionType.Field(i).Tag.Get("ini")
		// 如果和用户配置文件键名相同
		if filedName == key {
			filedKeyName = sectionType.Field(i).Name
			break
		}
	}

	// 获取字段值信息
	filedVal := sectionVal.FieldByName(filedKeyName)
	k := filedVal.Type().Kind()

	switch k {
	case reflect.String:
		// 设置字段值
		filedVal.SetString(val)
		return nil
	case reflect.Int:
		val, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			err = fmt.Errorf("syntax err,please check config file, must be a int64 number")
			return err
		}
		filedVal.SetInt(val)
		return nil

	default:
		err := fmt.Errorf("not support type")
		return err
	}
}

// 序列化配置文件
func Marshal(i interface{}) ([]byte, error) {
	v := reflect.ValueOf(i)
	t := v.Type()

	if t.Kind() != reflect.Struct {
		err := errors.New("please pass struct")
		return nil, err
	}
	var result []byte
	var data []string

	// 遍历节点
	for i := 0; i < t.NumField(); i++ {
		sectionVal := v.Field(i)
		sectionType := sectionVal.Type()

		// 如果节点的类型不是结构体，直接报错
		if sectionType.Kind() != reflect.Struct {
			continue
		}

		// 获取节点名称
		sectionName := t.Field(i).Tag.Get("ini")

		if len(sectionName) == 0 {
			sectionName = t.Field(i).Name
		}

		sectionName = fmt.Sprintf("[%s]\n", sectionName)
		data = append(data, sectionName)

		// 循环获取字段
		for j := 0; j < sectionType.NumField(); j++ {

			// 获取字段名
			filedName := sectionType.Field(j).Tag.Get("ini")
			// 如果没有ini tag，直接使用字段名
			if len(filedName) == 0 {
				filedName = sectionType.Field(i).Name
			}

			filedVal := sectionVal.Field(j).Interface()
			filed := fmt.Sprintf("%s = %v\n", filedName, filedVal)

			data = append(data, filed)
		}
	}

	for _, v := range data {
		byteVal := []byte(v)
		result = append(result, byteVal...)

	}

	return result, nil

}
